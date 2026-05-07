package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zalberix/cactus/apps/workers/telegram/config"
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

func telegramVariants(cfg config.Config) map[string]worker.Variant {
	return map[string]worker.Variant{
		"basic": {
			Name:     "basic",
			Manifest: basicTelegramManifest(),
			Handler:  &TelegramHandler{serverURL: cfg.ServerURL},
		},
	}
}

func basicTelegramManifest() worker.ManifestSpec {
	return worker.Manifest().
		Kind("telegram-basic", "Telegram Bot").
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
		Build()
}
