package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"

	dto "cactus/internal/DTO"
	"cactus/internal/logger"
	configschema "cactus/internal/pkg/configSchema"
	rdb "cactus/internal/storage/redis"
	"cactus/lib/worker"
)

type TelegramWorkerConfig struct {
	ServerURL  string `validate:"required,url" slug:"server_url"`
	worker     *worker.Worker
	mutex      *sync.Mutex
	sendChan   chan func()
	stopWorker chan struct{}
}

func NewTelegramWorkerConfig(worker *worker.Worker) *TelegramWorkerConfig {
	smtpWorker := &TelegramWorkerConfig{
		worker:     worker,
		mutex:      &sync.Mutex{},
		sendChan:   make(chan func()),
		stopWorker: make(chan struct{}),
	}
	smtpWorker.run()
	return smtpWorker
}

func (conf *TelegramWorkerConfig) run() {
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

func (conf *TelegramWorkerConfig) Stop() {
	close(conf.stopWorker)
}

func (conf *TelegramWorkerConfig) Update(values map[string]interface{}) {
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

func (conf *TelegramWorkerConfig) Send(message dto.MessageValueInMessageQueue, _ dto.SystemValueInMessageQueue) {
	conf.mutex.Lock()
	defer conf.mutex.Unlock()

	conf.sendChan <- func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var msgContent struct {
			Text string `json:"message"`
		}
		if err := json.Unmarshal(message.Value, &msgContent); err != nil {
			conf.worker.Err() <- fmt.Errorf("ошибка парсинга сообщения: %w", err)
			return
		}
		slog.Info("message", "message", msgContent)

		requestBody, err := json.Marshal(map[string]string{
			"message": msgContent.Text,
		})
		if err != nil {
			conf.worker.Err() <- fmt.Errorf("ошибка формирования запроса: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, conf.ServerURL, bytes.NewReader(requestBody))
		if err != nil {
			conf.worker.Err() <- fmt.Errorf("ошибка создания запроса: %w", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			conf.worker.Err() <- fmt.Errorf("ошибка выполнения запроса: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			conf.worker.Err() <- fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
		}
	}
}

func main() {
	Type := "social"
	NameType := "Рассылка через соц. сети"
	Kind := "telegram"
	NameKind := "Телеграм бот"

	conf := MustLoad()

	if conf.WorkerUUID == "" {
		slog.Error("worker обязан иметь ID (UUID)")
		return
	}

	ctx := context.Background()

	slog.SetDefault(logger.SetupLogger(conf.Env))

	RDBStorage, err := rdb.New(ctx, &redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		Username: conf.Redis.User,
	})
	if err != nil {
		slog.Error("Ошибка создания соединения с redis: ", slog.Any("error", err.Error()))
		return
	}

	defer func() {
		err := RDBStorage.Close()
		if err != nil {
			slog.Error("Ошибка закрытия соединения с redis:", slog.Any("error", err.Error()))
		}
	}()

	broker := worker.NewBrokerRedis(RDBStorage)

	workerCore := worker.NewWorker(ctx, broker, worker.Config{
		Token:          conf.Token,
		WorkerKind:     Kind,
		WorkerNameKind: NameKind,
		WorkerType:     Type,
		WorkerNameType: NameType,
		WorkerUUID:     conf.WorkerUUID,
		ConfigSchema: []configschema.ConfigField{
			{
				Type: "text",
				Slug: "server_url",
				Name: "URL сервера для отправки сообщений.",
			},
		},
	})

	TelegramWorker := NewTelegramWorkerConfig(workerCore)

	workerCore.SetConfigHandler(func(message worker.Message) {
		TelegramWorker.Update(message.Value)
		fmt.Printf("Обновляем: %+v\n", TelegramWorker)
	})

	workerCore.SetHandler(func(m worker.QueueMessage) {
		defer m.Ack()

		// TODO вынести эту логику парсинга в worker этим не должен пользователь заниматься

		fmt.Printf("Отправляю: %+v\n", m.Message.UUID)
		TelegramWorker.Send(m.Message, m.System)
	})

	fmt.Println("Воркер телеграм запущен")

	// TODO добавить CTRL+C сигнал и graceful-shutdown
	workerCore.Run()

	fmt.Println("Воркер телеграм остановлен")
}
