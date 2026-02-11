package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"cactus/pkg/configschema"
	"cactus/pkg/contracts"
	"cactus/pkg/pipeline"
)

type Message struct {
	ID string
	// TODO сделать generic тип для Value сообщения, реализовать можно через struct tags.
	Value map[string]interface{}
}

type workerMeta struct {
	// TODO сделать tag meta для того чтобы понимать какое поле надо искать в meta event.
	MaxPriority      int    `slug:"max_priority"`
	RegisterEndpoint string `slug:"register_endpoint"`
}

type QueueMessage struct {
	stream   *StreamConfig
	ID       string
	Pipeline contracts.PipelineValueInMessageQueue `json:"pipeline"`
	Message  contracts.MessageValueInMessageQueue  `json:"message"`
	System   contracts.SystemValueInMessageQueue   `json:"system"`
}

func (m *QueueMessage) Ack() error {
	err := m.stream.worker.broker.Ack(m.stream.worker.ctx, m.stream.Name, m.stream.worker.groupName, m.ID)
	if err != nil {
		m.stream.worker.err <- err
	}
	return err
}

type ArgStream struct {
	Name  string
	ID    string
	Block time.Duration
	Count int64
}

type StreamConfig struct {
	worker *Worker
	ArgStream
}

type Worker struct {
	ctx           context.Context
	broker        Broker
	config        Config
	groupName     string
	streamsEvent  map[string]ArgStream
	configHandler func(Message)
	handler       func(QueueMessage)
	err           chan error
	logger        func(error)
	meta          workerMeta
	ready         chan struct{}
}

type Config struct {
	Token          string
	WorkerKind     string
	WorkerNameKind string
	WorkerType     string
	WorkerNameType string
	WorkerUUID     string
	ConfigSchema []configschema.ConfigField
}

type Broker interface {
	Ack(ctx context.Context, name string, groupName string, id string) error
	Read(ctx context.Context, streams []string, block time.Duration, count int64) (BrokerMessages, error)
	ReadGroup(
		ctx context.Context, group string, consumer string,
		streams []string, block time.Duration, count int64,
	) (BrokerMessages, error)
	Add(ctx context.Context, stream string, id string, values map[string]interface{}) error
	ReadLatestMessages(ctx context.Context, stream string, count int64) ([]BrokerMessage, error)
	CreateStreamIfNotExist(ctx context.Context, key string, group string) error
}

func NewWorker(
	ctx context.Context,
	broker Broker,
	config Config,
) *Worker {
	return &Worker{
		ctx:    ctx,
		broker: broker,
		config: config,

		// TODO возможно нужно вынести, но как будто бы пользователь ни чего не должен знать
		// о том откуда он получает сообщения, пока на этапе до MVP не понятно.
		groupName: "reader",
		streamsEvent: map[string]ArgStream{
			"config": {
				Name:  "event:config:" + config.WorkerKind,
				ID:    "$",
				Block: 0,
				Count: 1,
			},
			"meta": {
				Name:  "event:meta",
				ID:    "$",
				Block: 0,
				Count: 1,
			},
		},
		configHandler: func(_ Message) { slog.Info("configHandler по умолчанию") },
		logger: func(err error) {
			slog.Error("ошибка в работе воркера:", slog.String("error", err.Error()))
		},
		err:   make(chan error),
		ready: make(chan struct{}),
	}
}

func (w *Worker) log() {
	for e := range w.err {
		w.logger(e)
	}
}

func (w *Worker) Err() chan<- error {
	return w.err
}

func (w *Worker) SetConfigStream(stream ArgStream) {
	w.streamsEvent["config"] = stream
}

// Пока не понятно нужно ли давать возможность добавлять слушать дополнительные стримы,
// так как по сути стримы сейчас стандартизированные. И нет возможности указать другие.
// func (w *Worker) AddQueueStream(stream ArgStream) {
// 	w.streamsTask = append(w.streamsTask, stream)
// }

func (w *Worker) SetConfigHandler(handler func(Message)) {
	w.configHandler = handler
}

func (w *Worker) SetHandler(handler func(QueueMessage)) {
	w.handler = handler
}

func (w *Worker) Run() {
	// Запуск логера
	go w.log()

	if w.handler == nil {
		w.err <- errors.New("функция для обработки сообщений является обязательной")
		return
	}

	err := w.readInitMetaStream(w.streamsEvent["meta"].Name)
	if err != nil {
		w.err <- err
	}
	// slog.Info("w.meta", "w.meta", w.meta)

	err = w.registerWorker()
	if err != nil {
		w.err <- err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func(arg ArgStream) {
		defer wg.Done()
		w.readEventStream(
			w.ctx,
			StreamConfig{
				worker:    w,
				ArgStream: arg,
			}, func(m Message) {
				w.configHandler(m)

				if !w.IsReady() {
					close(w.ready)
				}
			},
		)
	}(w.streamsEvent["config"])

	wg.Add(1)
	go func(arg ArgStream) {
		defer wg.Done()
		w.readEventStream(
			w.ctx,
			StreamConfig{
				worker:    w,
				ArgStream: arg,
			}, func(message Message) {
				err = MapToStruct(message.Value, &w.meta)
				if err != nil {
					w.err <- fmt.Errorf("невозможно получить значения из сообщения стрима %v: %w", message.Value, err)
				}
			},
		)
	}(w.streamsEvent["meta"])

	for weight := 0; weight <= w.meta.MaxPriority*2; weight++ {
		wg.Add(1)
		go func(weight int, workerType string, workerKind string) {
			defer wg.Done()

			streamName := fmt.Sprintf("messages:%v:%v:w-%v", workerType, workerKind, weight)

			<-w.ready

			ctx, cancel := context.WithCancel(w.ctx)
			defer cancel()

			w.readQueueStream(ctx, StreamConfig{
				worker: w,
				ArgStream: ArgStream{
					Name:  streamName,
					ID:    ">",
					Block: 10 * time.Second,
					Count: 5,
				},
			}, w.handler)
		}(weight, w.config.WorkerType, w.config.WorkerKind)
	}

	wg.Wait()
}

func (w *Worker) readEventStream(ctx context.Context, stream StreamConfig, handler func(Message)) {
	err := w.createStreamIfNotExist(stream.Name, "")
	if err != nil {
		w.err <- fmt.Errorf("ошибка создания стрима %v: %w", stream.Name, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("ctx выполнен")
			return
		default:
			msgs, err := w.broker.Read(w.ctx, []string{stream.Name, stream.ID}, stream.Block, stream.Count)
			if err != nil && !errors.Is(err, redis.Nil) {
				w.err <- fmt.Errorf("ошибка чтения стрима %v: %w", stream.Name, err)
				continue
			}

			for _, msg := range msgs {
				for _, m := range msg {
					handler(Message{
						ID:    m.ID,
						Value: m.Values,
					})
				}
			}
		}
	}
}

// Чтение из стрима Redis и обработка через handler
//
// TODO сделать чтобы возвращался канал, а не принимался handler,
// так как при изменении количества приоритетов (meta MaxPriority) нужно
// чтобы была возможность дочитать сообщения и вернуть их в сервис,
// a еще как то надо обыграть block 0 либо сделать block 1 мин. Надо перфоманс посмотреть.
// Тут сложнее чем кажется, завершение ctx не будет работать для XReadGroup если Block = 0
// # https://github.com/redis/go-redis/issues/2556
// Тоесть завершить горутину можно будет только если XReadGroup вычитали сообщение,
// В случае если стрим больше не нужен и туда не пишутся сообщения,
// он не сомжет завершиться, так как нет сообщений для чтения.
func (w *Worker) readQueueStream(ctx context.Context, stream StreamConfig, handler func(QueueMessage)) {
	err := w.createStreamIfNotExist(stream.Name, stream.worker.groupName)
	if err != nil {
		w.err <- fmt.Errorf("ошибка создания стрима %v: %w", stream.Name, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("ctx выполнен")
			return
		default:
			msgs, err := w.broker.ReadGroup(
				ctx,
				stream.worker.groupName,
				stream.worker.config.WorkerUUID,
				[]string{stream.Name, stream.ID},
				stream.Block,
				stream.Count,
			)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				if !errors.Is(err, redis.Nil) {
					w.err <- fmt.Errorf("ошибка чтения стрима %v: %w", stream.Name, err)
				}
				continue
			}

			for _, msg := range msgs {
				for _, m := range msg {
					queueMessage, err := w.ParseValueMessage(m.ID, m.Values, stream)
					if err != nil {
						w.Err() <- err
						continue
					}

					w.sendStatusWorkFor(queueMessage)
					handler(queueMessage)
					w.sendStatusDoneFor(queueMessage)
				}
			}
		}
	}
}

func (w *Worker) ParseValueMessage(
	id string,
	values map[string]interface{},
	stream StreamConfig,
) (QueueMessage, error) {
	var pipelineVal contracts.PipelineValueInMessageQueue
	err := w.GetFromValue(values, "pipeline", &pipelineVal)
	if err != nil {
		return QueueMessage{}, err
	}

	var system contracts.SystemValueInMessageQueue
	err = w.GetFromValue(values, "system", &system)
	if err != nil {
		return QueueMessage{}, err
	}

	var message contracts.MessageValueInMessageQueue
	err = w.GetFromValue(values, "message", &message)
	if err != nil {
		return QueueMessage{}, err
	}

	queueMessage := QueueMessage{
		stream:   &stream,
		ID:       id,
		Pipeline: pipelineVal,
		System:   system,
		Message:  message,
	}

	return queueMessage, nil
}

func (w *Worker) createStreamIfNotExist(key string, group string) error {
	return w.broker.CreateStreamIfNotExist(w.ctx, key, group)
}

type responseSuccess struct {
	Success bool `json:"success"`
}

type responseData[T any] struct {
	Data T `json:"data"`
}

func (w *Worker) registerWorker() error {
	if w.meta.RegisterEndpoint == "" {
		return fmt.Errorf("невозможно зарегистрировать воркер без endpoint для его регистрации")
	}
	ctx := w.ctx

	requestRegisterEndpoint := contracts.RegisterWorkerRequest{
		Token:        w.config.Token,
		WorkerUUID:   w.config.WorkerUUID,
		Kind:         w.config.WorkerKind,
		NameKind:     w.config.WorkerNameKind,
		Type:         w.config.WorkerType,
		NameType:     w.config.WorkerNameType,
		ConfigSchema: w.config.ConfigSchema,
	}

	requestRegisterEndpointJSON, err := json.Marshal(requestRegisterEndpoint)
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON для запроса в endpoint для регистрации worker: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		w.meta.RegisterEndpoint,
		bytes.NewBuffer(requestRegisterEndpointJSON),
	)
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("ошибка чтения тела ответа: %w", readErr)
		}
		bodyString := string(bodyBytes)

		return fmt.Errorf(
			"ошибка ответа от сервера: статус %v (%v) answer: %v",
			resp.Status,
			string(requestRegisterEndpointJSON),
			bodyString,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	var successResp responseSuccess

	if err := json.Unmarshal(body, &successResp); err != nil {
		return fmt.Errorf("ошибка десериализации JSON-ответа: %w", err)
	}

	if !successResp.Success {
		return fmt.Errorf("регистрация невозможна: %+v", string(body))
	}

	var dataResp responseData[contracts.RegisterWorkerResponse]

	if err := json.Unmarshal(body, &dataResp); err != nil {
		return fmt.Errorf("ошибка десериализации JSON-ответа: %w", err)
	}

	if len(dataResp.Data.Config) != 0 {
		w.configHandler(Message{
			ID:    "-",
			Value: dataResp.Data.Config,
		})
		if !w.IsReady() {
			close(w.ready)
		}
	}

	return nil
}

func (w *Worker) readInitMetaStream(stream string) error {
	messages, err := w.broker.ReadLatestMessages(w.ctx, stream, 1)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return fmt.Errorf("нет сообщений в стриме")
	}
	lastMessage := messages[0]

	err = MapToStruct(lastMessage.Values, &w.meta)
	if err != nil {
		return fmt.Errorf("невозможно получить значения из сообщения стрима %v: %w", lastMessage.Values, err)
	}

	return nil
}

func (w *Worker) IsReady() bool {
	select {
	case <-w.ready:
		return true
	default:
	}
	return false
}

func (w *Worker) GetFromValue(values map[string]interface{}, key string, target any) error {
	JSONInt, ok := values[key]
	if !ok {
		return fmt.Errorf("в сообщении отсутсвуют данные по ключу %v", key)
	}

	JSON, ok := JSONInt.(string)
	if !ok {
		return fmt.Errorf("сообщение не строка по ключу %v", key)
	}

	if err := json.Unmarshal([]byte(JSON), &target); err != nil {
		return fmt.Errorf("сообщение по ключу %v не валидный JSON", key)
	}
	return nil
}

func (w *Worker) sendStatusWorkFor(m QueueMessage) {
	JSONm, err := json.Marshal(pipeline.PipelineMessage{
		Status:    pipeline.Work,
		Step:      m.Pipeline.Step,
		WorkeUUID: w.config.WorkerUUID,
		UUID:      m.Message.UUID.String(),
	})
	if err != nil {
		w.err <- err
		return
	}

	err = w.broker.Add(w.ctx, "event:pipeline", "*",
		map[string]interface{}{
			"message": JSONm,
		},
	)
	if err != nil {
		w.err <- err
		return
	}
}

func (w *Worker) sendStatusDoneFor(m QueueMessage) {
	JSONm, err := json.Marshal(pipeline.PipelineMessage{
		Status:    pipeline.Done,
		Step:      m.Pipeline.Step,
		WorkeUUID: w.config.WorkerUUID,
		UUID:      m.Message.UUID.String(),
	})
	if err != nil {
		w.err <- err
		return
	}

	err = w.broker.Add(w.ctx, "event:pipeline", "*",
		map[string]interface{}{
			"message": JSONm,
		},
	)
	if err != nil {
		w.err <- err
		return
	}
}

func MapToStruct(data map[string]interface{}, result interface{}) error {
	resultValue := reflect.ValueOf(result).Elem()
	resultType := reflect.TypeOf(result).Elem()

	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)

		metaTag, ok := field.Tag.Lookup("slug")
		if !ok {
			continue
		}

		value, found := data[metaTag]
		if !found {
			continue
		}

		fieldValue := resultValue.FieldByName(field.Name)
		if fieldValue.IsValid() && fieldValue.CanSet() {
			switch fieldValue.Kind() { //nolint:exhaustive
			case reflect.String:
				strValue, ok := value.(string)
				if !ok {
					return fmt.Errorf("поле '%s' должно быть строкой", field.Name)
				}
				fieldValue.SetString(strValue)
			case reflect.Int:
				strValue, ok := value.(string)
				if !ok {
					return fmt.Errorf("поле '%s' должно быть строкой", field.Name)
				}

				intValue, err := strconv.ParseInt(strValue, 10, 64)
				if err != nil {
					return fmt.Errorf("поле '%s' должно быть числом (%w)", field.Name, err)
				}
				fieldValue.SetInt(intValue)
				// TODO: Добавить обработку других типов
				// линтер отключил так как default не считается за обработку всех типов
			default:
				return fmt.Errorf("неподдерживаемый тип для поля '%s'", field.Name)
			}
		} else {
			return fmt.Errorf("поле '%s' недоступно для записи", field.Name)
		}
	}
	return nil
}
