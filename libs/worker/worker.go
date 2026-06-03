package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	defaultTaskTimeout       = 2 * time.Minute
	defaultAckWait           = 30 * time.Second
	defaultMaxDeliver        = 5
	defaultConsumeRetryDelay = time.Second
	defaultFetchMaxWait      = 2 * time.Second
	defaultNATSReconnectWait = time.Second
)

func taskFilterSubject(orgID, workTypeID, settingsSchemaID int32) string {
	return fmt.Sprintf("task.org.%d.work_type.%d.schema.%d.>", orgID, workTypeID, settingsSchemaID)
}

func taskConsumerName(orgID, workTypeID, settingsSchemaID int32) string {
	return fmt.Sprintf("task-org-%d-work-type-%d-schema-%d", orgID, workTypeID, settingsSchemaID)
}

func taskConsumerConfig(orgID, workTypeID, settingsSchemaID int32) jetstream.ConsumerConfig {
	name := taskConsumerName(orgID, workTypeID, settingsSchemaID)
	return jetstream.ConsumerConfig{
		Name:          name,
		Durable:       name,
		FilterSubject: taskFilterSubject(orgID, workTypeID, settingsSchemaID),
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		MaxDeliver:    defaultMaxDeliver,
		AckWait:       defaultAckWait,
	}
}

// Worker is the SDK runtime for NATS JetStream-backed task workers.
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

// New creates a Worker SDK instance. It does not connect to NATS until Run.
func New(cfg Config, handler TaskHandler, logger *slog.Logger) *Worker {
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = defaultHeartbeatInterval
	}
	if cfg.TaskTimeout <= 0 {
		cfg.TaskTimeout = defaultTaskTimeout
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

// Run registers the worker, connects to NATS, starts heartbeat, and consumes tasks until ctx is canceled.
func (w *Worker) Run(ctx context.Context) error {
	if err := w.registerWithRetry(ctx); err != nil {
		return fmt.Errorf("registration: %w", err)
	}
	if err := w.requireRoutingConfigured(); err != nil {
		return err
	}

	go w.heartbeatLoop(ctx)

	return w.natsLoop(ctx)
}

func (w *Worker) natsLoop(ctx context.Context) error {
	for {
		if err := w.connectNATS(); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			w.logger.Warn("connect to NATS failed, retrying",
				slog.String("error", err.Error()),
				slog.Duration("retry_in", defaultNATSReconnectWait),
			)
			if err := waitForRetry(ctx, defaultNATSReconnectWait); err != nil {
				return err
			}
			continue
		}

		w.logger.Info("worker registered",
			slog.Int("worker_id", int(w.workerID)),
			slog.String("name", w.cfg.WorkerName),
		)

		err := w.consumeLoop(ctx)
		if ctx.Err() != nil {
			w.drainNATS()
			return ctx.Err()
		}

		w.closeNATS()
		w.logger.Warn("NATS connection lost, reconnecting",
			slog.String("error", err.Error()),
			slog.Duration("retry_in", defaultNATSReconnectWait),
		)
		if err := waitForRetry(ctx, defaultNATSReconnectWait); err != nil {
			return err
		}
	}
}

func (w *Worker) connectNATS() error {
	opts := []nats.Option{
		nats.UserJWTAndSeed(w.natsCreds.UserJWT, w.natsCreds.UserSeed),
	}
	opts = append(opts, w.natsConnectOptions()...)
	nc, err := nats.Connect(w.natsCreds.URL, opts...)
	if err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}
	w.nc = nc

	js, err := jetstream.New(nc)
	if err != nil {
		w.nc.Close()
		w.nc = nil
		return fmt.Errorf("init JetStream: %w", err)
	}
	w.js = js

	return nil
}

func (w *Worker) natsConnectOptions() []nats.Option {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(defaultNATSReconnectWait),
	}
	if w.natsCreds.CAFile != "" {
		opts = append(opts, nats.RootCAs(w.natsCreds.CAFile))
	} else if w.cfg.NatsCAFile != "" {
		opts = append(opts, nats.RootCAs(w.cfg.NatsCAFile))
	}
	return opts
}

func (w *Worker) drainNATS() {
	if w.nc == nil {
		return
	}
	_ = w.nc.Drain()
	w.nc = nil
	w.js = nil
}

func (w *Worker) closeNATS() {
	if w.nc == nil {
		return
	}
	w.nc.Close()
	w.nc = nil
	w.js = nil
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

func (w *Worker) refreshLoggerAfterWorkerID() {
	if w.workerID <= 0 || w.cfg.OnWorkerID == nil {
		return
	}
	if logger := w.cfg.OnWorkerID(w.workerID); logger != nil {
		w.logger = logger
	}
}

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

func (w *Worker) consumeLoop(ctx context.Context) error {
	for {
		if err := w.consumeOnce(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if isNATSConnectionError(err) {
				return err
			}
			w.logger.Warn("task consumer stopped, recreating",
				slog.String("error", err.Error()),
				slog.Duration("retry_in", defaultConsumeRetryDelay),
			)
			if err := waitForRetry(ctx, defaultConsumeRetryDelay); err != nil {
				return err
			}
			continue
		}
		return nil
	}
}

func (w *Worker) consumeOnce(ctx context.Context) error {
	subject := taskFilterSubject(w.cfg.OrganizationID, w.cfg.WorkTypeID, w.cfg.WorkerSettingsSchemaID)
	consumerName := taskConsumerName(w.cfg.OrganizationID, w.cfg.WorkTypeID, w.cfg.WorkerSettingsSchemaID)

	cons, err := w.js.CreateOrUpdateConsumer(
		ctx,
		"TASKS",
		taskConsumerConfig(w.cfg.OrganizationID, w.cfg.WorkTypeID, w.cfg.WorkerSettingsSchemaID),
	)
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	w.logger.Info("consuming tasks",
		slog.String("subject", subject),
		slog.String("consumer", consumerName),
		slog.Int("worker_id", int(w.workerID)),
		slog.Int("settings_schema_id", int(w.cfg.WorkerSettingsSchemaID)),
	)

	return w.consume(ctx, cons)
}

func (w *Worker) consume(ctx context.Context, cons jetstream.Consumer) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		batch, err := cons.Fetch(1, jetstream.FetchMaxWait(defaultFetchMaxWait))
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			if isNATSConnectionError(err) {
				return err
			}
			w.logger.Error("fetching task failed", slog.String("error", err.Error()))
			continue
		}

		for msg := range batch.Messages() {
			w.processMessage(ctx, msg)
		}

		if err := batch.Error(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrTimeout) {
				continue
			}
			if isNATSConnectionError(err) {
				return err
			}
			w.logger.Error("task fetch batch failed", slog.String("error", err.Error()))
		}
	}
}

func isNATSConnectionError(err error) bool {
	return errors.Is(err, nats.ErrConnectionClosed) || errors.Is(err, nats.ErrDisconnected)
}

func (w *Worker) processMessage(ctx context.Context, msg jetstream.Msg) {
	w.logger.Info(
		"received task message",
		append([]any{"subject", msg.Subject()}, taskMsgMeta(msg)...)...,
	)

	taskCtx, cancel := context.WithTimeout(ctx, w.cfg.TaskTimeout)
	defer cancel()

	var task TaskMessage
	if err := json.Unmarshal(msg.Data(), &task); err != nil {
		w.logger.Error("unmarshal task",
			slog.String("error", err.Error()),
			slog.String("subject", msg.Subject()),
		)
		if nakErr := msg.Nak(); nakErr != nil {
			w.logger.Error("nacking malformed task failed", slog.String("error", nakErr.Error()))
		}
		return
	}

	w.logger.Info("resolving task settings",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.String("config_subject", task.ConfigRef.ConfigSubject),
	)

	settings, err := w.resolveSettings(taskCtx, task.ConfigRef)
	if err != nil {
		result := Result{Success: false, Error: err.Error(), WorkerID: w.workerID}
		if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
			result.Error = fmt.Sprintf("worker task timed out after %s", w.cfg.TaskTimeout)
		}
		w.publishAndAck(resultPublishContext(ctx, taskCtx), msg, task, result)
		return
	}
	task.Settings = settings

	w.logger.Info("resolved task settings",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
	)
	w.logger.Info("handling task",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
	)

	result, err := w.handler.Handle(taskCtx, task)
	if err != nil {
		result = Result{
			Success:  false,
			Error:    err.Error(),
			WorkerID: w.workerID,
		}
	} else {
		result.WorkerID = w.workerID
	}
	if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
		result = Result{
			Success:  false,
			Error:    fmt.Sprintf("worker task timed out after %s", w.cfg.TaskTimeout),
			WorkerID: w.workerID,
		}
	}

	w.logger.Info("handled task",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
		slog.Bool("success", result.Success),
	)

	w.publishAndAck(resultPublishContext(ctx, taskCtx), msg, task, result)
}

func resultPublishContext(parentCtx, taskCtx context.Context) context.Context {
	if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
		return parentCtx
	}
	return taskCtx
}

func (w *Worker) publishAndAck(ctx context.Context, msg jetstream.Msg, task TaskMessage, result Result) {
	if err := w.publishResult(ctx, task.ReplyTo, result); err != nil {
		w.logger.Error("publishing task result failed",
			append(taskLogAttrs(task), "reply_to", task.ReplyTo, "error", err.Error())...,
		)
		if nakErr := msg.Nak(); nakErr != nil {
			w.logger.Error("nacking task after result publish failure failed",
				append(taskLogAttrs(task), "error", nakErr.Error())...,
			)
		}
		return
	}

	w.logger.Info("published task result",
		append(taskLogAttrs(task), "reply_to", task.ReplyTo)...,
	)

	if err := msg.Ack(); err != nil {
		w.logger.Error("acking task after result publish failed",
			append(append(taskLogAttrs(task), taskMsgMeta(msg)...), "error", err.Error())...,
		)
		return
	}

	w.logger.Info("acked task",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
	)
}

func (w *Worker) publishResult(ctx context.Context, replyTo string, result Result) error {
	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return fmt.Errorf("marshal result: %w", marshalErr)
	}

	if _, pubErr := w.js.Publish(ctx, replyTo, resultJSON); pubErr != nil {
		return fmt.Errorf("publish result: %w", pubErr)
	}

	return nil
}

func taskLogAttrs(task TaskMessage) []any {
	return []any{
		"workflow_run_id", int(task.WorkflowRunID),
		"step_id", int(task.StepID),
		"attempt", int(task.Attempt),
	}
}

func taskMsgMeta(msg jetstream.Msg) []any {
	meta, err := msg.Metadata()
	if err != nil {
		return nil
	}

	return []any{
		"stream_seq", meta.Sequence.Stream,
		"consumer_seq", meta.Sequence.Consumer,
		"num_delivered", meta.NumDelivered,
		"num_pending", meta.NumPending,
	}
}
