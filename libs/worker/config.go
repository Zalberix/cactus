package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nats-io/nats.go"
)

// fetchConfig получает текущую конфигурацию ревизии от Manager через HTTP.
// GET {ManagerURL}/api/v1/workers/{workerID}/config
// Header: X-Bootstrap-Token: {bootstrapToken}
func (w *Worker) fetchConfig(ctx context.Context) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/v1/workers/%d/config",
		strings.TrimRight(w.cfg.ManagerURL, "/"),
		w.workerID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create config request: %w", err)
	}
	req.Header.Set("X-Bootstrap-Token", w.cfg.BootstrapToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send config request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read config response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("config fetch failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var envelope struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal config response: %w", err)
	}

	if !envelope.Success {
		return nil, fmt.Errorf("config fetch rejected: %s", string(body))
	}

	return envelope.Data, nil
}

// subscribeConfigReload подписывается на NATS subject event.config.{manifest.Kind}
// для hot-reload уведомлений. При получении сообщения --- запрашивает свежий конфиг.
func (w *Worker) subscribeConfigReload(ctx context.Context) error {
	subject := "event.config." + w.cfg.Manifest.Kind

	_, err := w.nc.Subscribe(subject, func(_ *nats.Msg) {
		w.logger.Info("config reload notification received",
			slog.String("subject", subject),
		)

		cfg, fetchErr := w.fetchConfig(ctx)
		if fetchErr != nil {
			w.logger.Error("config reload fetch failed",
				slog.String("error", fetchErr.Error()),
			)
			return
		}

		w.logger.Info("config reloaded successfully",
			slog.Any("config", cfg),
		)
	})
	if err != nil {
		return fmt.Errorf("subscribe config reload on %s: %w", subject, err)
	}

	w.logger.Info("subscribed to config reload",
		slog.String("subject", subject),
	)

	return nil
}
