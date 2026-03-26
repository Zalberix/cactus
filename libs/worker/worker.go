package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	defaultAckWait           = 30 * time.Second
	defaultMaxDeliver        = 5
	defaultInactiveThreshold = 5 * time.Minute
)

// Worker --- SDK для подключения воркеров к Manager через NATS JetStream.
type Worker struct {
	cfg      Config
	handler  TaskHandler
	nc       *nats.Conn
	js       jetstream.JetStream
	workerID int32
	logger   *slog.Logger
}

// New создаёт Worker SDK instance. Не подключается к NATS до вызова Run.
func New(cfg Config, handler TaskHandler, logger *slog.Logger) *Worker {
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = defaultHeartbeatInterval
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{
		cfg:     cfg,
		handler: handler,
		logger:  logger,
	}
}

// Run подключается к NATS, регистрируется в Manager, запускает heartbeat
// и начинает обработку задач. Блокирует до отмены ctx.
// Graceful shutdown: останавливает consumer, heartbeat, drains NATS.
func (w *Worker) Run(ctx context.Context) error {
	// 1. Подключение к NATS
	nc, err := nats.Connect(w.cfg.NatsURL)
	if err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}
	w.nc = nc

	js, err := jetstream.New(nc)
	if err != nil {
		w.nc.Close()
		return fmt.Errorf("init JetStream: %w", err)
	}
	w.js = js

	defer func() {
		_ = w.nc.Drain()
	}()

	// 2. Регистрация или загрузка workerID
	if err := w.loadOrRegister(ctx); err != nil {
		return fmt.Errorf("registration: %w", err)
	}

	w.logger.Info("worker registered",
		slog.Int("worker_id", int(w.workerID)),
		slog.String("name", w.cfg.WorkerName),
	)

	// 3. Подписка на config reload
	if err := w.subscribeConfigReload(ctx); err != nil {
		w.logger.Warn("config reload subscription failed", slog.String("error", err.Error()))
	}

	// 4. Запуск heartbeat
	go w.heartbeatLoop(ctx)

	// 5. Consume loop
	return w.consumeLoop(ctx)
}

// consumeLoop подписывается на TASKS stream и обрабатывает сообщения.
func (w *Worker) consumeLoop(ctx context.Context) error {
	subject := fmt.Sprintf("task.%d.%d.>", w.cfg.WorkTypeID, w.cfg.RevisionID)
	consumerName := fmt.Sprintf("worker-%d", w.workerID)

	cons, err := w.js.CreateOrUpdateConsumer(ctx, "TASKS", jetstream.ConsumerConfig{
		Durable:           consumerName,
		FilterSubject:     subject,
		AckPolicy:         jetstream.AckExplicitPolicy,
		DeliverPolicy:     jetstream.DeliverAllPolicy,
		MaxDeliver:        defaultMaxDeliver,
		AckWait:           defaultAckWait,
		InactiveThreshold: defaultInactiveThreshold,
	})
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	w.logger.Info("consuming tasks",
		slog.String("subject", subject),
		slog.String("consumer", consumerName),
	)

	iter, err := cons.Messages()
	if err != nil {
		return fmt.Errorf("create message iterator: %w", err)
	}
	defer iter.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := iter.Next()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			w.logger.Error("fetch next message", slog.String("error", err.Error()))
			continue
		}

		w.processMessage(ctx, msg)
	}
}

// processMessage обрабатывает одно сообщение из NATS.
func (w *Worker) processMessage(ctx context.Context, msg jetstream.Msg) {
	var task TaskMessage
	if err := json.Unmarshal(msg.Data(), &task); err != nil {
		w.logger.Error("unmarshal task",
			slog.String("error", err.Error()),
			slog.String("subject", msg.Subject()),
		)
		_ = msg.Nak()
		return
	}

	w.logger.Info("processing task",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
	)

	// Вызов пользовательского handler
	result, err := w.handler.Handle(ctx, task)
	if err != nil {
		result = Result{
			Success:  false,
			Error:    err.Error(),
			WorkerID: w.workerID,
		}
	} else {
		result.WorkerID = w.workerID
	}

	// Публикация результата в ReplyTo subject (RESULTS stream)
	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		w.logger.Error("marshal result", slog.String("error", marshalErr.Error()))
		_ = msg.Nak()
		return
	}

	if _, pubErr := w.js.Publish(ctx, task.ReplyTo, resultJSON); pubErr != nil {
		w.logger.Error("publish result",
			slog.String("reply_to", task.ReplyTo),
			slog.String("error", pubErr.Error()),
		)
		_ = msg.Nak()
		return
	}

	_ = msg.Ack()

	w.logger.Info("task completed",
		slog.Int("step_id", int(task.StepID)),
		slog.Bool("success", result.Success),
	)
}
