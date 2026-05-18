package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

// TaskHandler --- user-implemented business logic (per D-02).
// Каждый воркер (SMTP, Telegram и т.д.) реализует этот интерфейс.
type TaskHandler interface {
	Handle(ctx context.Context, task TaskMessage) (Result, error)
}

// TaskMessage --- задача от Manager через NATS.
// JSON-wire-compatible с temporal.TaskMessage (apps/core/internal/temporal/types.go).
type TaskMessage struct {
	WorkflowRunID  int32          `json:"workflow_run_id"`
	StepID         int32          `json:"step_id"`
	Attempt        int32          `json:"attempt"`
	ReplyTo        string         `json:"reply_to"`
	Input          map[string]any `json:"input"`
	ConfigRef      ConfigRef      `json:"config_ref"`
	Settings       map[string]any `json:"-"`
	IdempotencyKey string         `json:"idempotency_key"`
}

type ConfigRef struct {
	OrganizationID int32  `json:"organization_id"`
	WorkTypeID     int32  `json:"work_type_id"`
	SchemaID       int32  `json:"schema_id"`
	RevisionID     int32  `json:"revision_id"`
	ConfigHash     string `json:"config_hash"`
	ConfigSubject  string `json:"config_subject"`
}

type NATSCredentials struct {
	URL         string `json:"url"`
	CAFile      string `json:"ca_file,omitempty"`
	UserJWT     string `json:"user_jwt"`
	UserSeed    string `json:"user_seed"`
	Credentials string `json:"credentials"`
}

// Result --- результат выполнения задачи, публикуется в ReplyTo subject.
// JSON-wire-compatible с temporal.WorkerResult.
type Result struct {
	Success  bool           `json:"success"`
	Output   map[string]any `json:"output,omitempty"`
	Error    string         `json:"error,omitempty"`
	WorkerID int32          `json:"worker_id,omitempty"`
}

// Config --- конфигурация Worker SDK.
type Config struct {
	ManagerURL             string                            `json:"manager_url"`
	BootstrapToken         string                            `json:"bootstrap_token"`
	NatsCAFile             string                            `json:"nats_ca_file"`
	OrganizationID         int32                             `json:"organization_id"`
	WorkTypeID             int32                             `json:"work_type_id"`
	WorkerSettingsSchemaID int32                             `json:"worker_settings_schema_id"`
	RevisionID             int32                             `json:"revision_id"`
	WorkerIDPath           string                            `json:"worker_id_file"`
	WorkerName             string                            `json:"worker_name"`
	HeartbeatInterval      time.Duration                     `json:"heartbeat_interval"`
	Manifest               ManifestSpec                      `json:"manifest"`
	OnWorkerID             func(workerID int32) *slog.Logger `json:"-"`
}

// ManifestSpec --- возможности воркера, объявляемые при регистрации (per D-03).
type ManifestSpec struct {
	Kind           string          `json:"kind"`
	NameKind       string          `json:"name_kind"`
	Type           string          `json:"type"`
	NameType       string          `json:"name_type"`
	SettingsSchema json.RawMessage `json:"settings_schema"`
	InputSchema    json.RawMessage `json:"input_schema"`
	OutputSchema   json.RawMessage `json:"output_schema"`
}
