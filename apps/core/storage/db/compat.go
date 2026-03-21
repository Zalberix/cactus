package db

// Backward-compatibility types for code that still uses the old schema (v0.2.0).
// These types exist ONLY in Go — no corresponding tables in the DB.
// TODO: remove once all callers migrate to the new schema (v0.3.0 — Phase 2).

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// ─── Old types (no DB tables) ───────────────────────────────────────────────

type TypeWorker struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type KindWorker struct {
	ID           int32                 `json:"id"`
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	ConfigSchema json.RawMessage       `json:"config_schema"`
	Config       pqtype.NullRawMessage `json:"config"`
}

type OldPipeline struct {
	ID               int32         `json:"id"`
	MessageID        int32         `json:"message_id"`
	ParentPipelineID sql.NullInt32 `json:"parent_pipeline_id"`
	CreatedAt        sql.NullTime  `json:"created_at"`
	UpdatedAt        sql.NullTime  `json:"updated_at"`
	DeletedAt        sql.NullTime  `json:"deleted_at"`
}

// Pipeline is an alias kept for old code compatibility.
type Pipeline = OldPipeline

type OldPipelineStep struct {
	ID                   int32         `json:"id"`
	PipelineID           int32         `json:"pipeline_id"`
	WorkerID             sql.NullInt32 `json:"worker_id"`
	ChannelID            int32         `json:"channel_id"`
	Step                 int32         `json:"step"`
	TimeStart            sql.NullTime  `json:"time_start"`
	TimeEnd              sql.NullTime  `json:"time_end"`
	CreatedAt            sql.NullTime  `json:"created_at"`
	UpdatedAt            sql.NullTime  `json:"updated_at"`
	DeletedAt            sql.NullTime  `json:"deleted_at"`
	PipelineStepStatusID int32         `json:"pipeline_step_status_id"`
}

type PipelineStep = OldPipelineStep

type PipelineStepStatus struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Priority struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	Weight int32  `json:"weight"`
	Slug   string `json:"slug"`
}

type GetPriorityBySystemIDRow struct {
	Weight int32 `json:"weight"`
}

type Token struct {
	IDSystem     int32  `json:"id_system"`
	IDKindWorker int32  `json:"id_kind_worker"`
	IsActive     bool   `json:"is_active"`
	PublicToken  string `json:"public_token"`
	SecretToken  string `json:"secret_token"`
}

type Manifest struct {
	ID    int32           `json:"id"`
	Value json.RawMessage `json:"value"`
}

type Channel struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Config struct {
	ID           int32                 `json:"id"`
	Name         string                `json:"name"`
	ConfigSchema json.RawMessage       `json:"config_schema"`
	Config       pqtype.NullRawMessage `json:"config"`
}

// ─── Old param types ────────────────────────────────────────────────────────

type CreateChannelParams struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type CreateTypeWorkerParams = CreateChannelParams

type CreateKindWorkerParams struct {
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	ConfigSchema json.RawMessage       `json:"config_schema"`
	Config       pqtype.NullRawMessage `json:"config"`
}

type CreateConfigParams struct {
	Name         string                `json:"name"`
	ConfigSchema json.RawMessage       `json:"config_schema"`
	Config       pqtype.NullRawMessage `json:"config"`
}

type CreateManifestParams struct {
	Value json.RawMessage `json:"value"`
}

type CreatePipelineParams struct {
	MessageID        int32         `json:"message_id"`
	ParentPipelineID sql.NullInt32 `json:"parent_pipeline_id"`
}

type CreatePipelineStepParams struct {
	PipelineID           int32         `json:"pipeline_id"`
	WorkerID             sql.NullInt32 `json:"worker_id"`
	ChannelID            int32         `json:"channel_id"`
	Step                 int32         `json:"step"`
	TimeStart            sql.NullTime  `json:"time_start"`
	TimeEnd              sql.NullTime  `json:"time_end"`
	PipelineStepStatusID int32         `json:"pipeline_step_status_id"`
}

type CreatePipelineStepStatusParams struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdatePipelineStepStatusAndWorkerParams struct {
	ID                   int32         `json:"id"`
	PipelineStepStatusID int32         `json:"pipeline_step_status_id"`
	WorkerID             sql.NullInt32 `json:"worker_id"`
}

type UpdatePipelineStatusAndWorkerByIDParams struct {
	ID       int32         `json:"id"`
	Status   string        `json:"status"`
	IDWorker sql.NullInt32 `json:"id_worker"`
}

type GetIDPipelineByUUIDMessageAndStepParams struct {
	UUID uuid.UUID `json:"uuid"`
	Step int32     `json:"step"`
}

type GetPipelineStepByPipelineIDAndStepParams struct {
	PipelineID int32 `json:"pipeline_id"`
	Step       int32 `json:"step"`
}

type GetMessagesByParams struct {
	IDTypeWorker int32 `json:"id_type_worker"`
	IDSystem     int32 `json:"id_system"`
}

// Old param types kept for backward compat — service layer uses these until Phase 2 rewrite.

type CreateFileParams struct {
	MessageID int32  `json:"message_id"`
	Title     string `json:"title"`
	Name      string `json:"name"`
	Ext       string `json:"ext"`
	Url       string `json:"url"`
}

type CreateMessageParams struct {
	SystemID   int32           `json:"system_id"`
	ManifestID int32           `json:"manifest_id"`
	Uuid       uuid.UUID       `json:"uuid"`
	Priority   int32           `json:"priority"`
	Value      json.RawMessage `json:"value"`
	SendAt     sql.NullTime    `json:"send_at"`
}

type CreateWorkerParams struct {
	ChannelID int32 `json:"channel_id"`
	ConfigID  int32 `json:"config_id"`
	IsActive  bool  `json:"is_active"`
}

type AddChannelForSystemParams struct {
	SystemID  int32 `json:"system_id"`
	ChannelID int32 `json:"channel_id"`
}
