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
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// registerRequest --- тело запроса на регистрацию воркера (POST /api/v1/register/worker).
// Matches apps/core/internal/domain/worktype.RegisterWorkerRequest.
type registerRequest struct {
	BootstrapToken string          `json:"bootstrap_token"`
	Name           string          `json:"name"`
	Manifest       json.RawMessage `json:"manifest"`
}

// registerResponse --- envelope ответа от Manager.
// Data contains db.Worker fields; we only parse what we need.
type registerResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID     int32          `json:"id"`
		Config map[string]any `json:"config,omitempty"`
	} `json:"data"`
}

// loadOrRegister загружает workerID из файла или регистрирует новый.
// При наличии файла с workerID отправляет heartbeat для подтверждения.
// При отсутствии --- регистрируется через POST /api/v1/register/worker.
// Сохраняет workerID в файл.
func (w *Worker) loadOrRegister(ctx context.Context) error {
	// Попытка загрузить workerID из файла
	if w.cfg.WorkerIDPath != "" {
		id, err := w.loadWorkerID()
		if err == nil && id > 0 {
			w.workerID = id
			w.logger.Info("loaded worker ID from file",
				slog.Int("worker_id", int(id)),
				slog.String("file", w.cfg.WorkerIDPath),
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
	if w.cfg.WorkerIDPath != "" && w.workerID > 0 {
		if err := w.saveWorkerID(w.workerID); err != nil {
			w.logger.Warn("failed to save worker ID to file",
				slog.String("error", err.Error()),
				slog.String("file", w.cfg.WorkerIDPath),
			)
		}
	}

	return nil
}

// sendRegistration отправляет POST /api/v1/register/worker.
func (w *Worker) sendRegistration(ctx context.Context) error {
	manifestJSON, err := json.Marshal(w.cfg.Manifest)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	reqBody := registerRequest{
		BootstrapToken: w.cfg.BootstrapToken,
		Name:           w.cfg.WorkerName,
		Manifest:       manifestJSON,
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

	if regResp.Data.ID > 0 {
		w.workerID = regResp.Data.ID
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
	data, err := os.ReadFile(w.cfg.WorkerIDPath)
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
// Создаёт директорию если она не существует.
func (w *Worker) saveWorkerID(id int32) error {
	dir := filepath.Dir(w.cfg.WorkerIDPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create worker ID dir %s: %w", dir, err)
	}
	return os.WriteFile(w.cfg.WorkerIDPath, []byte(strconv.Itoa(int(id))), 0644)
}
