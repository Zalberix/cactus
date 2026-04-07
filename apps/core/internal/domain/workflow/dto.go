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
	StepType                 string          `json:"step_type" binding:"required,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
	ControlKind              *string         `json:"control_kind"`
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"`
	CanvasPosition           json.RawMessage `json:"canvas_position"`
}

// UpdateStepRequest — запрос на обновление шага.
type UpdateStepRequest struct {
	StepType                 string          `json:"step_type" binding:"required,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
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
}

// ValidateVersionResponse — ответ на запрос валидации версии.
type ValidateVersionResponse struct {
	IsValid bool                    `json:"is_valid"`
	Errors  []dagpkg.ValidationError `json:"errors,omitempty"`
}
