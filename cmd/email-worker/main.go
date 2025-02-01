package main

import (
	dto "cactus/internal/DTO"
	"cactus/internal/logger"
	configschema "cactus/internal/pkg/configSchema"
	"cactus/internal/plugin/email"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	rdb "cactus/internal/storage/redis"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

type SMTPWorkerConfig struct {
	Host       string `validate:"required,ip" slug:"host"`
	Port       int    `validate:"required,numeric" slug:"port"`
	From       string `validate:"required,email" slug:"from"`
	worker     *Worker
	mutex      *sync.Mutex
	sendChan   chan func()
	stopWorker chan struct{}
}

func NewSMTPWorkerConfig(worker *Worker) *SMTPWorkerConfig {
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

	MapToStruct(values, conf)

	validate := validator.New()
	if err := validate.Struct(conf); err != nil {
		slog.Info("Ошибка валидации при обновлении", slog.Any("err", err.Error()))
		*conf = backup
	}
}

func (conf *SMTPWorkerConfig) Send(message dto.MessageValueInMessageQueue, system dto.SystemValueInMessageQueue) {
	conf.mutex.Lock()
	defer conf.mutex.Unlock()
	conf.sendChan <- func() {

		var SMTPValue email.EmailSchema
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
				resp, err := http.Get(file.Url)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					// TODO обработать сообщение
					return fmt.Errorf("ошибка загрузки файла: статус %d", resp.StatusCode)
				}

				_, err = io.Copy(w, resp.Body)
				return err
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

	worker := NewWorker(ctx, RDBStorage, WorkerConfig{
		Token:      conf.Token,
		WorkerKind: Kind,
		WorkerType: Type,
		WorkerUUID: conf.WorkerUUID,
		ConfigSchema: []configschema.ConfigField{
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

	SMTPWorker := NewSMTPWorkerConfig(worker)

	worker.SetConfigHandler(func(message Message) {
		SMTPWorker.Update(message.Value)
		fmt.Printf("Обновляем: %+v\n", SMTPWorker)
	})

	worker.SetHandler(func(m QueueMessage) {
		defer m.Act()

		time.Sleep(1 * time.Second)

		systemJsonInt, ok := m.Value["system"]
		if !ok {
			worker.Err() <- fmt.Errorf("отсутсвуют данные системы. Сообщение будет пропущено")
			// TODO сделать вывод error в event:error чтобы основной сервис подхватывал
			return
		}

		systemJson, ok := systemJsonInt.(string)
		if !ok {
			worker.Err() <- fmt.Errorf("отсутсвуют данные системы. Сообщение будет пропущено")
			// TODO сделать вывод error в event:error чтобы основной сервис подхватывал
			return
		}

		var system dto.SystemValueInMessageQueue
		if err := json.Unmarshal([]byte(systemJson), &system); err != nil {
			worker.Err() <- fmt.Errorf("отсутсвует информация о системем которая отправила сообщение. Сообщение будет пропущено")
			// TODO сделать вывод error в event:error чтобы основной сервис подхватывал
			return
		}

		messageJsonInt, ok := m.Value["message"]
		if !ok {
			worker.Err() <- fmt.Errorf("отсутсвуют данные сообщения. Сообщение будет пропущено")
			// TODO сделать вывод error в event:error чтобы основной сервис подхватывал
			return
		}

		messageJson, ok := messageJsonInt.(string)
		if !ok {
			worker.Err() <- fmt.Errorf("отсутсвуют данные сообщения. Сообщение будет пропущено")
			// TODO сделать вывод error в event:error чтобы основной сервис подхватывал
			return
		}

		var message dto.MessageValueInMessageQueue
		if err = json.Unmarshal([]byte(messageJson), &message); err != nil {
			worker.Err() <- fmt.Errorf("отсутсвуют данные о системем которая отправила сообщение. Сообщение будет пропущено")
			// TODO сделать вывод error в event:error чтобы основной сервис подхватывал
			return

		}

		fmt.Printf("Отправляю: %+v\n", message)
		SMTPWorker.Send(message, system)
	})

	fmt.Println("Воркер запущен")

	// TODO добавить CTRL+C сигнал и graceful-shutdown
	worker.Run()

	fmt.Println("Воркер остановлен")
}
