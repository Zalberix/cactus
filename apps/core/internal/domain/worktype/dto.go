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

// WorkerResponse — воркер с вычисленным статусом.
type WorkerResponse struct {
	ID              int32        `json:"id"`
	Name            string       `json:"name"`
	WorkTypeID      int32        `json:"work_type_id"`
	SchemaID        int32        `json:"schema_id"`
	Status          WorkerStatus `json:"status"`
	LastHeartbeatAt time.Time    `json:"last_heartbeat_at"`
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
