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

func taskFilterSubject(orgID, workTypeID int32) string {
	return fmt.Sprintf("task.org.%d.work_type.%d.>", orgID, workTypeID)
}

// Worker --- SDK для подключения воркеров к Manager через NATS JetStream.
type Worker struct {
	cfg         Config
	handler     TaskHandler
	nc          *nats.Conn
	js          jetstream.JetStream
	workerID    int32
	logger      *slog.Logger
	natsCreds   NATSCredentials
	configCache *configCache
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
		cfg:         cfg,
		handler:     handler,
		logger:      logger,
		configCache: newConfigCache(),
	}
}

// Run подключается к NATS, регистрируется в Manager, запускает heartbeat
// и начинает обработку задач. Блокирует до отмены ctx.
// Graceful shutdown: останавливает consumer, heartbeat, drains NATS.
func (w *Worker) Run(ctx context.Context) error {
	// 1. Подключение к NATS
	if err := w.registerWithRetry(ctx); err != nil {
		return fmt.Errorf("registration: %w", err)
	}
	if err := w.requireRoutingConfigured(); err != nil {
		return err
	}

	opts := []nats.Option{
		nats.UserJWTAndSeed(w.natsCreds.UserJWT, w.natsCreds.UserSeed),
	}
	if w.natsCreds.CAFile != "" {
		opts = append(opts, nats.RootCAs(w.natsCreds.CAFile))
	} else if w.cfg.NatsCAFile != "" {
		opts = append(opts, nats.RootCAs(w.cfg.NatsCAFile))
	}
	nc, err := nats.Connect(w.natsCreds.URL, opts...)
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

	// 2. Регистрация или загрузка workerID (с retry)
	w.logger.Info("worker registered",
		slog.Int("worker_id", int(w.workerID)),
		slog.String("name", w.cfg.WorkerName),
	)

	// 3. Подписка на config reload
	// 4. Запуск heartbeat
	go w.heartbeatLoop(ctx)

	// 5. Consume loop
	return w.consumeLoop(ctx)
}

func (w *Worker) refreshLoggerAfterWorkerID() {
	if w.workerID <= 0 || w.cfg.OnWorkerID == nil {
		return
	}
	if logger := w.cfg.OnWorkerID(w.workerID); logger != nil {
		w.logger = logger
	}
}

// registerWithRetry пытается зарегистрироваться с экспоненциальным backoff.
// Не сдаётся до отмены ctx.
func (w *Worker) requireRoutingConfigured() error {
	if w.cfg.OrganizationID <= 0 || w.cfg.WorkTypeID <= 0 || w.cfg.WorkerSettingsSchemaID <= 0 || w.cfg.RevisionID <= 0 {
		return fmt.Errorf("registration response missing organization_id, work_type_id, worker_settings_schema_id or revision_id")
	}
	if w.natsCreds.URL == "" || w.natsCreds.UserJWT == "" || w.natsCreds.UserSeed == "" {
		return fmt.Errorf("registration response missing nats credentials")
	}
	return nil
}

func (w *Worker) registerWithRetry(ctx context.Context) error {
	backoff := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
	}
	const maxDelay = 5 * time.Second

	for attempt := 0; ; attempt++ {
		err := w.loadOrRegister(ctx)
		if err == nil {
			return nil
		}

		delay := maxDelay
		if attempt < len(backoff) {
			delay = backoff[attempt]
		}

		w.logger.Warn("registration failed, retrying...",
			slog.String("error", err.Error()),
			slog.Int("attempt", attempt+1),
			slog.Duration("retry_in", delay),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

// consumeLoop подписывается на TASKS stream и обрабатывает сообщения.
func (w *Worker) consumeLoop(ctx context.Context) error {
	subject := taskFilterSubject(w.cfg.OrganizationID, w.cfg.WorkTypeID)
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
	settings, err := w.resolveSettings(ctx, task.ConfigRef)
	if err != nil {
		w.publishResult(ctx, task.ReplyTo, Result{Success: false, Error: err.Error(), WorkerID: w.workerID}, msg)
		return
	}
	task.Settings = settings

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
	w.publishResult(ctx, task.ReplyTo, result, msg)
}

func (w *Worker) publishResult(ctx context.Context, replyTo string, result Result, msg jetstream.Msg) {
	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		w.logger.Error("marshal result", slog.String("error", marshalErr.Error()))
		_ = msg.Nak()
		return
	}

	if _, pubErr := w.js.Publish(ctx, replyTo, resultJSON); pubErr != nil {
		w.logger.Error("publish result",
			slog.String("reply_to", replyTo),
			slog.String("error", pubErr.Error()),
		)
		_ = msg.Nak()
		return
	}

	_ = msg.Ack()

	w.logger.Info("task completed",
		slog.Bool("success", result.Success),
	)
}
