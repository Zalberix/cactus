package worktype

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// WorkerStatus — статус воркера.
type WorkerStatus string

const (
	WorkerStatusWorking WorkerStatus = "working"
	WorkerStatusReady   WorkerStatus = "ready"
	WorkerStatusOffline WorkerStatus = "offline"
)

// --- Work Type DTOs ---

// CreateWorkTypeRequest — запрос на создание типа работы.
type CreateWorkTypeRequest struct {
	Name        string          `json:"name" binding:"required,min=2,max=255"`
	Code        string          `json:"code" binding:"required,min=2,max=255"`
	Description string          `json:"description"`
	Meta        json.RawMessage `json:"meta"`
}

// Response — DTO для API, meta как json.RawMessage (не base64).
type Response struct {
	ID          int32           `json:"id"`
	Name        string          `json:"name"`
	Code        string          `json:"code"`
	Description string          `json:"description,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

type WorkTypeCatalogItem struct {
	ID           int32                 `json:"id"`
	Name         string                `json:"name"`
	Code         string                `json:"code"`
	Description  string                `json:"description,omitempty"`
	Meta         json.RawMessage       `json:"meta,omitempty"`
	WorkerCount  int64                 `json:"worker_count"`
	ReadyWorkers int64                 `json:"ready_workers"`
	Schemas      []SettingsSchemaBrief `json:"schemas"`
}

type SettingsSchemaBrief struct {
	ID             int32           `json:"id"`
	Version        string          `json:"version"`
	CreatedAt      string          `json:"created_at"`
	SettingsSchema json.RawMessage `json:"settings_schema"`
	WorkerCount    int64           `json:"worker_count"`
	ReadyWorkers   int64           `json:"ready_workers"`
}

// CreateWorkTypeResponse — ответ с типом работы и bootstrap-токеном.
type CreateWorkTypeResponse struct {
	WorkType       db.WorkType `json:"work_type"`
	BootstrapToken string      `json:"bootstrap_token"`
}

// --- Worker DTOs ---

// RegisterWorkerRequest — запрос регистрации воркера.
type RegisterWorkerRequest struct {
	BootstrapToken string          `json:"bootstrap_token" binding:"required"`
	Name           string          `json:"name" binding:"required"`
	Manifest       json.RawMessage `json:"manifest" binding:"required"`
}

type RegisterWorkerResponse struct {
	db.Worker
	RevisionID int32 `json:"revision_id"`
}

// WorkerResponse — воркер с вычисленным статусом.
type WorkerResponse struct {
	ID              int32        `json:"id"`
	Name            string       `json:"name"`
	WorkTypeID      int32        `json:"work_type_id"`
	SchemaID        int32        `json:"schema_id"`
	Status          WorkerStatus `json:"status"`
	LastHeartbeatAt time.Time    `json:"last_heartbeat_at"`
}

type DeleteWorkerResult string

const (
	DeleteWorkerResultDeleted  DeleteWorkerResult = "deleted"
	DeleteWorkerResultOnline   DeleteWorkerResult = "online"
	DeleteWorkerResultInUse    DeleteWorkerResult = "in_use"
	DeleteWorkerResultNotFound DeleteWorkerResult = "not_found"
)

type WorkerWorkflowUsage struct {
	WorkflowID            int32  `json:"workflow_id"`
	WorkflowName          string `json:"workflow_name"`
	SystemID              int32  `json:"system_id"`
	WorkflowVersionID     int32  `json:"workflow_version_id"`
	WorkflowVersionNumber int32  `json:"workflow_version_number"`
}

type DeleteWorkerResponse struct {
	Result   DeleteWorkerResult    `json:"result"`
	Deleted  bool                  `json:"deleted"`
	Message  string                `json:"message"`
	Usages   []WorkerWorkflowUsage `json:"usages,omitempty"`
	WorkerID int32                 `json:"worker_id,omitempty"`
	Status   WorkerStatus          `json:"status,omitempty"`
}

// --- System DTOs ---

// CreateSystemRequest — запрос на создание системы.
type CreateSystemRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description"`
	Priority    int32  `json:"priority"`
}

// UpdateSystemRequest — запрос на обновление системы.
type UpdateSystemRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// --- System Token DTOs ---

// CreateSystemTokenRequest — запрос на создание токена системы.
type CreateSystemTokenRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// CreateSystemTokenResponse — ответ с токенами системы (возвращается ОДИН раз).
type CreateSystemTokenResponse struct {
	ID           int32  `json:"id"`
	SystemID     int32  `json:"system_id"`
	Name         string `json:"name"`
	PublicToken  string `json:"public_token"`
	PrivateToken string `json:"private_token"`
}

type SystemTokenResponse struct {
	ID          int32            `json:"id"`
	SystemID    int32            `json:"system_id"`
	Name        string           `json:"name"`
	PublicToken string           `json:"public_token"`
	IsActive    bool             `json:"is_active"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at"`
}

// --- Settings Revision DTOs ---

// CreateRevisionRequest — запрос на создание ревизии настроек.
type CreateRevisionRequest struct {
	SettingsData json.RawMessage `json:"settings_data" binding:"required"`
}
