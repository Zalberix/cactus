package workflow

import (
	"encoding/json"

	dagpkg "github.com/zalberix/cactus/apps/core/internal/dag"
)

// CreateWorkflowRequest — запрос на создание workflow.
type CreateWorkflowRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Priority    int    `json:"priority" binding:"min=0,max=3"`
	Description string `json:"description"`
}

// UpdateWorkflowRequest — запрос на обновление workflow.
type UpdateWorkflowRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Priority    int    `json:"priority" binding:"min=0,max=3"`
	Description string `json:"description"`
}

// CreateStepRequest — запрос на создание шага.
type CreateStepRequest struct {
	Name                     string          `json:"name" binding:"omitempty,max=255"`
	StepType                 string          `json:"step_type" binding:"required,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
	WorkerSettingsSchemaID   *int32          `json:"worker_settings_schema_id"`
	ControlKind              *string         `json:"control_kind"`
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"`
	CanvasPosition           json.RawMessage `json:"canvas_position"`
}

// UpdateStepRequest — запрос на обновление шага.
type UpdateStepRequest struct {
	Name                     *string         `json:"name" binding:"omitempty,min=1,max=255"`
	StepType                 *string         `json:"step_type" binding:"omitempty,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
	WorkerSettingsSchemaID   *int32          `json:"worker_settings_schema_id"`
	ControlKind              *string         `json:"control_kind"`
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"`
	CanvasPosition           json.RawMessage `json:"canvas_position"`
}

// UpdateStepPositionRequest — запрос на обновление позиции шага на холсте.
type UpdateStepPositionRequest struct {
	CanvasPosition json.RawMessage `json:"canvas_position" binding:"required"`
}

// CreateDependencyRequest — запрос на создание зависимости между шагами.
type CreateDependencyRequest struct {
	DependsOnStepID int32  `json:"depends_on_step_id" binding:"required"`
	Outcome         string `json:"outcome" binding:"required"`
	OutputIndex     int32  `json:"output_index"`
}

// EnrichedStepResponse — enriched шаг для API ответа.
// Все []byte JSONB поля маппятся в json.RawMessage для корректной JSON-сериализации.
type EnrichedStepResponse struct {
	ID                       int32           `json:"id"`
	WorkflowVersionID        int32           `json:"workflow_version_id"`
	Name                     string          `json:"name"`
	StepType                 string          `json:"step_type"`
	WorkTypeID               *int32          `json:"work_type_id,omitempty"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id,omitempty"`
	ControlKind              *string         `json:"control_kind,omitempty"`
	ControlSettings          json.RawMessage `json:"control_settings,omitempty"`
	InputMapping             json.RawMessage `json:"input_mapping,omitempty"`
	CanvasPosition           json.RawMessage `json:"canvas_position,omitempty"`
	WorkTypeName             *string         `json:"work_type_name,omitempty"`
	WorkTypeCode             *string         `json:"work_type_code,omitempty"`
	WorkTypeMeta             json.RawMessage `json:"work_type_meta,omitempty"`
	SettingsSchema           json.RawMessage `json:"settings_schema,omitempty"`
	InputSchema              json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema             json.RawMessage `json:"output_schema,omitempty"`
	Config                   json.RawMessage `json:"config,omitempty"`
}

// ValidateVersionResponse — ответ на запрос валидации версии.
type ValidateVersionResponse struct {
	IsValid bool                     `json:"is_valid"`
	Errors  []dagpkg.ValidationError `json:"errors,omitempty"`
}

type VersionSummaryResponse struct {
	ID             int32  `json:"id"`
	WorkflowID     int32  `json:"workflow_id"`
	Name           string `json:"name"`
	VersionNumber  int32  `json:"version_number"`
	IsValid        bool   `json:"is_valid"`
	IsActive       bool   `json:"is_active"`
	TrafficWeight  int32  `json:"traffic_weight"`
	IsControlGroup bool   `json:"is_control_group"`
	RunCount       int64  `json:"run_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	DeletedAt      string `json:"deleted_at,omitempty"`
}

type UpdateVersionNameRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type UpdateTrafficRequest struct {
	Mode     string                `json:"mode" binding:"omitempty,oneof=equal custom"`
	Weights  []TrafficWeightInput  `json:"weights"`
	Versions []TrafficVersionInput `json:"versions"`
}

type TrafficWeightInput struct {
	VersionID int32 `json:"version_id" binding:"required"`
	Weight    int32 `json:"weight" binding:"min=0,max=100"`
}

type TrafficVersionInput struct {
	VersionID int32  `json:"version_id" binding:"required"`
	Mode      string `json:"mode" binding:"required,oneof=share fixed"`
	Weight    int32  `json:"weight" binding:"min=0,max=100"`
}

type InputSchemaFieldRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Type        string `json:"type" binding:"required,oneof=string number integer boolean object array"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type InputSchemaFieldResponse struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

type InputSchemaResponse struct {
	Schema map[string]any `json:"schema"`
}

type UpdateTaskSettingsRequest struct {
	SettingsData json.RawMessage `json:"settings_data" binding:"required"`
}

type UpdateTaskInputMappingRequest struct {
	InputMapping json.RawMessage `json:"input_mapping" binding:"required"`
}
