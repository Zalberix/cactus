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
	WorkTypeID               *int32          `json:"work_type_id"`                 // обязателен для task
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`  // обязателен для task
	ControlKind              *string         `json:"control_kind"`                 // обязателен для control
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"` // WF-05
}

// UpdateStepRequest — запрос на обновление шага.
type UpdateStepRequest struct {
	StepType                 string          `json:"step_type" binding:"required,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
	ControlKind              *string         `json:"control_kind"`
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"`
}

// CreateDependencyRequest — запрос на создание зависимости между шагами.
type CreateDependencyRequest struct {
	DependsOnStepID int32  `json:"depends_on_step_id" binding:"required"`
	Outcome         string `json:"outcome" binding:"required"`
}

// ValidateVersionResponse — ответ на запрос валидации версии.
type ValidateVersionResponse struct {
	IsValid bool                    `json:"is_valid"`
	Errors  []dagpkg.ValidationError `json:"errors,omitempty"`
}
