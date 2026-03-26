package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// registerRequest --- тело запроса на регистрацию воркера (POST /api/v1/register/worker).
type registerRequest struct {
	Token        string          `json:"token"`
	WorkerUUID   string          `json:"worker_uuid"`
	Kind         string          `json:"kind"`
	NameKind     string          `json:"name_kind"`
	Type         string          `json:"type"`
	NameType     string          `json:"name_type"`
	ConfigSchema json.RawMessage `json:"config_schema,omitempty"`
}

// registerResponse --- envelope ответа от Manager.
type registerResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Config map[string]any `json:"config,omitempty"`
	} `json:"data"`
}

// loadOrRegister загружает workerID из файла или регистрирует новый.
// При наличии файла с workerID отправляет heartbeat для подтверждения.
// При отсутствии --- регистрируется через POST /api/v1/register/worker.
// Сохраняет workerID в файл (per D-03).
func (w *Worker) loadOrRegister(ctx context.Context) error {
	// Попытка загрузить workerID из файла
	if w.cfg.WorkerIDFile != "" {
		id, err := w.loadWorkerID()
		if err == nil && id > 0 {
			w.workerID = id
			w.logger.Info("loaded worker ID from file",
				slog.Int("worker_id", int(id)),
				slog.String("file", w.cfg.WorkerIDFile),
			)
			// Отправляем heartbeat для подтверждения
			if err := w.sendRegistration(ctx); err != nil {
				w.logger.Warn("heartbeat after load failed, re-registering",
					slog.String("error", err.Error()),
				)
				// Если heartbeat не прошёл, регистрируемся заново
				return w.register(ctx)
			}
			return nil
		}
	}

	return w.register(ctx)
}

// register выполняет HTTP-регистрацию воркера и сохраняет workerID.
func (w *Worker) register(ctx context.Context) error {
	if err := w.sendRegistration(ctx); err != nil {
		return err
	}

	// Сохраняем workerID в файл
	if w.cfg.WorkerIDFile != "" && w.workerID > 0 {
		if err := w.saveWorkerID(w.workerID); err != nil {
			w.logger.Warn("failed to save worker ID to file",
				slog.String("error", err.Error()),
				slog.String("file", w.cfg.WorkerIDFile),
			)
		}
	}

	return nil
}

// sendRegistration отправляет POST /api/v1/register/worker.
func (w *Worker) sendRegistration(ctx context.Context) error {
	reqBody := registerRequest{
		Token:        w.cfg.BootstrapToken,
		WorkerUUID:   strconv.Itoa(int(w.workerID)),
		Kind:         w.cfg.Manifest.Kind,
		NameKind:     w.cfg.Manifest.NameKind,
		Type:         w.cfg.Manifest.Type,
		NameType:     w.cfg.Manifest.NameType,
		ConfigSchema: w.cfg.Manifest.InputSchema,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal registration request: %w", err)
	}

	url := strings.TrimRight(w.cfg.ManagerURL, "/") + "/api/v1/register/worker"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("create registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send registration request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read registration response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var regResp registerResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return fmt.Errorf("unmarshal registration response: %w", err)
	}

	if !regResp.Success {
		return fmt.Errorf("registration rejected: %s", string(respBody))
	}

	return nil
}

// heartbeatLoop отправляет периодический heartbeat (re-registration per D-05).
// Блокирует до отмены ctx.
func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.sendRegistration(ctx); err != nil {
				w.logger.Warn("heartbeat failed", slog.String("error", err.Error()))
			}
		}
	}
}

// loadWorkerID читает workerID из файла.
func (w *Worker) loadWorkerID() (int32, error) {
	data, err := os.ReadFile(w.cfg.WorkerIDFile)
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse worker ID from file: %w", err)
	}
	return int32(id), nil
}

// saveWorkerID записывает workerID в файл.
func (w *Worker) saveWorkerID(id int32) error {
	return os.WriteFile(w.cfg.WorkerIDFile, []byte(strconv.Itoa(int(id))), 0644)
}
