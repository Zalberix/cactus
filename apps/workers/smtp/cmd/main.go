package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"gopkg.in/gomail.v2"

	"github.com/zalberix/cactus/apps/workers/smtp/config"
	"github.com/zalberix/cactus/libs/logger"
	"github.com/zalberix/cactus/libs/pipeline"
	"github.com/zalberix/cactus/libs/worker"
)

// SMTPMessageSchema describes the JSON format of a message value for SMTP delivery.
type SMTPMessageSchema struct {
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Subject     string            `json:"subject"`
	TypeMessage *string           `json:"type_message"`
	Data        map[string]string `json:"data"`
}

type SMTPWorkerConfig struct {
	Host       string `validate:"required,ip" slug:"host"`
	Port       int    `validate:"required,numeric" slug:"port"`
	From       string `validate:"required,email" slug:"from"`
	worker     *worker.Worker
	mutex      *sync.Mutex
	sendChan   chan func()
	stopWorker chan struct{}
}

func NewSMTPWorkerConfig(worker *worker.Worker) *SMTPWorkerConfig {
	smtpWorker := &SMTPWorkerConfig{
		worker:     worker,
		mutex:      &sync.Mutex{},
		sendChan:   make(chan func()),
		stopWorker: make(chan struct{}),
	}
	smtpWorker.run()
	return smtpWorker
}

func (conf *SMTPWorkerConfig) run() {
	go func() {
		for {
			select {
			case task := <-conf.sendChan:
				task()
			case <-conf.stopWorker:
				return
			}
		}
	}()
}

func (conf *SMTPWorkerConfig) Stop() {
	close(conf.stopWorker)
}

func (conf *SMTPWorkerConfig) Update(values map[string]interface{}) {
	conf.mutex.Lock()
	defer conf.mutex.Unlock()

	backup := *conf

	worker.MapToStruct(values, conf)

	validate := validator.New()
	if err := validate.Struct(conf); err != nil {
		slog.Info("Ошибка валидации при обновлении", slog.Any("err", err.Error()))
		*conf = backup
	}
}

func (conf *SMTPWorkerConfig) Send(message pipeline.MessageInQueue, _ pipeline.SystemInQueue) {
	conf.mutex.Lock()
	defer conf.mutex.Unlock()
	conf.sendChan <- func() {
		var SMTPValue SMTPMessageSchema
		if err := json.Unmarshal(message.Value, &SMTPValue); err != nil {
			conf.worker.Err() <- fmt.Errorf("данные в value сообщения неверного формата: %w", err)
			return
		}

		d := gomail.Dialer{Host: conf.Host, Port: conf.Port}

		m := gomail.NewMessage()
		title := base64.StdEncoding.EncodeToString([]byte(SMTPValue.Title))
		m.SetHeader("From", conf.From)
		m.SetHeader("To", SMTPValue.Subject)
		m.SetHeader("Subject", fmt.Sprintf("=?UTF-8?B?%s?=", title))
		m.SetBody("text/html", SMTPValue.Message)

		for _, file := range message.Files {
			m.Attach("./dummy.txt", gomail.SetCopyFunc(func(w io.Writer) error {
				client := &http.Client{}
				req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, file.URL, nil)
				if err != nil {
					return fmt.Errorf("ошибка создания запроса: %w", err)
				}

				resp, err := client.Do(req)
				if err != nil {
					return fmt.Errorf("ошибка выполнения запроса: %w", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("ошибка загрузки файла: статус %d", resp.StatusCode)
				}

				_, err = io.Copy(w, resp.Body)
				if err != nil {
					return fmt.Errorf("ошибка копирования данных: %w", err)
				}

				return nil
			}), gomail.Rename(file.Name))
		}

		if err := d.DialAndSend(m); err != nil {
			conf.worker.Err() <- fmt.Errorf("ошибка отправки письма: %w", err)
		}
	}
}

func (conf *SMTPWorkerConfig) String() string {
	return fmt.Sprintf("address:%v:%v form:%v", conf.Host, conf.Port, conf.From)
}

func main() {
	Type := "email"
	Kind := "smtp"

	conf := config.MustLoad("./config/email.worker.yaml")

	if conf.WorkerUUID == "" {
		slog.Error("worker обязан иметь ID (UUID)")
		return
	}

	ctx := context.Background()

	slog.SetDefault(logger.SetupLogger(conf.Env))

	broker, err := worker.NewBrokerNATS(conf.Nats.URL)
	if err != nil {
		slog.Error("Ошибка создания соединения с NATS:", slog.Any("error", err.Error()))
		return
	}
	defer broker.Close()

	workerCore := worker.NewWorker(ctx, broker, worker.Config{
		Token:          conf.Token,
		WorkerKind:     Kind,
		WorkerNameKind: "SMTP рассылка",
		WorkerType:     Type,
		WorkerNameType: "Email рассылка",
		WorkerUUID:     conf.WorkerUUID,
		ConfigSchema: []pipeline.ConfigField{
			{
				Type: "host",
				Slug: "host",
				Name: "ip адрес сервера SMTP",
			},
			{
				Type: "numeric",
				Slug: "port",
				Name: "Порт сервера",
			},
			{
				Type: "email",
				Slug: "from",
				Name: "Адрем отправителя",
			},
		},
	})

	SMTPWorker := NewSMTPWorkerConfig(workerCore)

	workerCore.SetConfigHandler(func(message worker.Message) {
		SMTPWorker.Update(message.Value)
		fmt.Printf("Обновляем: %+v\n", SMTPWorker)
	})

	workerCore.SetHandler(func(m worker.QueueMessage) {
		defer m.Ack()

		fmt.Printf("Отправляю: %+v\n", m.Message.UUID)
		SMTPWorker.Send(m.Message, m.System)
	})
	slog.Info("Воркер smtp запущен")

	// TODO добавить CTRL+C сигнал и graceful-shutdown
	workerCore.Run()

	slog.Info("Воркер smtp остановлен")
}
