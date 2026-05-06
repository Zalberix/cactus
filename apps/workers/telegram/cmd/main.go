package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zalberix/cactus/apps/workers/telegram/config"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
	"github.com/zalberix/cactus/libs/worker"
)

type TelegramHandler struct {
	serverURL string
}

func (h *TelegramHandler) Handle(ctx context.Context, task worker.TaskMessage) (worker.Result, error) {
	text, _ := task.Input["message"].(string)
	if text == "" {
		text, _ = task.Input["text"].(string)
	}
	if text == "" {
		return worker.Result{}, fmt.Errorf("missing required field: message")
	}

	requestBody, err := json.Marshal(map[string]string{
		"message": text,
	})
	if err != nil {
		return worker.Result{}, fmt.Errorf("marshal request body: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, h.serverURL, bytes.NewReader(requestBody))
	if err != nil {
		return worker.Result{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return worker.Result{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return worker.Result{}, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	return worker.Result{
		Success: true,
		Output: map[string]any{
			"sent_at": time.Now().Format(time.RFC3339),
		},
	}, nil
}

func main() {
	workerIDPath := flag.String("worker-id-path", "", "path to worker ID file (.worker_id/{type}/{uuid})")
	flag.Parse()

	conf := cfgloader.MustLoad[config.Config]("configs/apps/workers/telegram.yaml")
	slog.SetDefault(logger.SetupLogger(conf.Env, logger.WorkerSource("telegram", 0)))

	bootstrapToken := conf.BootstrapToken
	if bootstrapToken == "" {
		bootstrapToken = conf.Token
	}
	if bootstrapToken == "" {
		slog.Error("bootstrap token is required")
		os.Exit(1)
	}

	handler := &TelegramHandler{serverURL: conf.ServerURL}
	workerCore := worker.New(worker.Config{
		NatsURL:        conf.NatsURL,
		ManagerURL:     conf.ManagerURL,
		BootstrapToken: bootstrapToken,
		WorkTypeID:     conf.WorkTypeID,
		RevisionID:     conf.RevisionID,
		WorkerIDPath:   *workerIDPath,
		WorkerName:     "telegram-worker",
		OnWorkerID: func(workerID int32) *slog.Logger {
			return logger.SetupLogger(conf.Env, logger.WorkerSource("telegram", workerID))
		},
		Manifest: worker.Manifest().
			Kind("telegram", "Telegram Bot").
			Type("social", "Social Delivery").
			SettingsSchema(func(sb *worker.SchemaBuilder) {
				sb.String("server_url").Required()
			}).
			InputSchema(func(sb *worker.SchemaBuilder) {
				sb.String("message").Required()
			}).
			OutputSchema(func(sb *worker.SchemaBuilder) {
				sb.String("sent_at")
			}).
			Build(),
	}, handler, slog.Default())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("Telegram worker starting")

	if err := workerCore.Run(ctx); err != nil {
		slog.Error("worker stopped", slog.String("error", err.Error()))
		cancel()
		return
	}
}
