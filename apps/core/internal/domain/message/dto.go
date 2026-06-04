package message

import "time"

// SendMessageRequest is the inbound message request.
type SendMessageRequest struct {
	WorkflowID         int32          `json:"workflow_id" binding:"required"`
	InputSchemaID      *int32         `json:"input_schema_id,omitempty"`
	InputSchemaCode    string         `json:"input_schema_code,omitempty"`
	IdempotencyKey     string         `json:"idempotency_key,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	OverriddenPriority *int32         `json:"overridden_priority,omitempty"`
	Value              map[string]any `json:"value" binding:"required"`
	ExternalID         string         `json:"external_id"`
}

// SendMessageResponse is returned after a message is accepted and routed.
type SendMessageResponse struct {
	MessageID                        int32  `json:"message_id"`
	WorkflowRunID                    int32  `json:"workflow_run_id"`
	WorkflowInputSchemaID            *int32 `json:"workflow_input_schema_id,omitempty"`
	WorkflowInputSchemaCode          string `json:"workflow_input_schema_code,omitempty"`
	WorkflowInputSchemaVersionNumber *int32 `json:"workflow_input_schema_version_number,omitempty"`
	WorkflowVersionID                *int32 `json:"workflow_version_id,omitempty"`
	InputSchemaCompatibilityID       *int32 `json:"input_schema_compatibility_id,omitempty"`
	SelectionReason                  string `json:"selection_reason,omitempty"`
	WorkflowExperimentID             *int32 `json:"workflow_experiment_id,omitempty"`
	WorkflowExperimentScopeID        *int32 `json:"workflow_experiment_scope_id,omitempty"`
	WorkflowExperimentVariantID      *int32 `json:"workflow_experiment_variant_id,omitempty"`
	Status                           string `json:"status"`
}

// StatusResponse is the Status API envelope.
type StatusResponse struct {
	MessageID     int32              `json:"message_id"`
	MessageStatus string             `json:"message_status"`
	CreatedAt     time.Time          `json:"created_at"`
	WorkflowRun   *WorkflowRunStatus `json:"workflow_run,omitempty"`
	Steps         []StepStatusDTO    `json:"steps"`
}

// WorkflowRunStatus is nested workflow_run info in status response.
type WorkflowRunStatus struct {
	ID           int32      `json:"id"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	ErrorMessage *string    `json:"error_message"`
}

// StepStatusDTO is step status without input/output.
type StepStatusDTO struct {
	ID           int32      `json:"id"`
	StepID       int32      `json:"step_id"`
	StepType     string     `json:"step_type"`
	Status       string     `json:"status"`
	Outcome      *string    `json:"outcome"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
	ErrorMessage *string    `json:"error_message"`
}

// ListItem is a single message in organization listing.
type ListItem struct {
	ID                               int32   `json:"id"`
	WorkflowID                       int32   `json:"workflow_id"`
	WorkflowName                     string  `json:"workflow_name"`
	WorkflowInputSchemaID            *int32  `json:"workflow_input_schema_id,omitempty"`
	WorkflowInputSchemaCode          *string `json:"workflow_input_schema_code,omitempty"`
	WorkflowInputSchemaVersionNumber *int32  `json:"workflow_input_schema_version_number,omitempty"`
	WorkflowVersionID                *int32  `json:"workflow_version_id,omitempty"`
	WorkflowVersionNumber            *int32  `json:"workflow_version_number,omitempty"`
	WorkflowVersionName              *string `json:"workflow_version_name,omitempty"`
	WorkflowExperimentID             *int32  `json:"workflow_experiment_id,omitempty"`
	WorkflowExperimentVariantID      *int32  `json:"workflow_experiment_variant_id,omitempty"`
	SelectionReason                  *string `json:"selection_reason,omitempty"`
	Status                           string  `json:"status"`
	CreatedAt                        string  `json:"created_at"`
	UpdatedAt                        string  `json:"updated_at"`
}

type DetailResponse struct {
	MessageID                        int32              `json:"message_id"`
	WorkflowID                       int32              `json:"workflow_id"`
	WorkflowName                     string             `json:"workflow_name"`
	WorkflowInputSchemaID            *int32             `json:"workflow_input_schema_id,omitempty"`
	WorkflowInputSchemaCode          *string            `json:"workflow_input_schema_code,omitempty"`
	WorkflowInputSchemaVersionNumber *int32             `json:"workflow_input_schema_version_number,omitempty"`
	WorkflowVersionID                *int32             `json:"workflow_version_id,omitempty"`
	WorkflowVersionNumber            *int32             `json:"workflow_version_number,omitempty"`
	WorkflowVersionName              *string            `json:"workflow_version_name,omitempty"`
	WorkflowExperimentID             *int32             `json:"workflow_experiment_id,omitempty"`
	WorkflowExperimentScopeID        *int32             `json:"workflow_experiment_scope_id,omitempty"`
	WorkflowExperimentVariantID      *int32             `json:"workflow_experiment_variant_id,omitempty"`
	InputSchemaCompatibilityID       *int32             `json:"input_schema_compatibility_id,omitempty"`
	SelectionReason                  *string            `json:"selection_reason,omitempty"`
	MessageStatus                    string             `json:"message_status"`
	MessageValue                     map[string]any     `json:"message_value"`
	MessageMetadata                  map[string]any     `json:"message_metadata,omitempty"`
	MessageErrorMessage              *string            `json:"message_error_message,omitempty"`
	VersionInputData                 map[string]any     `json:"version_input_data,omitempty"`
	RoutingDecision                  map[string]any     `json:"routing_decision,omitempty"`
	CreatedAt                        time.Time          `json:"created_at"`
	UpdatedAt                        time.Time          `json:"updated_at"`
	WorkflowRun                      *WorkflowRunStatus `json:"workflow_run,omitempty"`
	Graph                            GraphDTO           `json:"graph"`
	RunSteps                         []StepRunDetailDTO `json:"run_steps"`
}

type GraphDTO struct {
	VersionID    int32                `json:"version_id"`
	Steps        []GraphStepDTO       `json:"steps"`
	Dependencies []GraphDependencyDTO `json:"dependencies"`
}

type GraphStepDTO struct {
	ID             int32          `json:"id"`
	StepType       string         `json:"step_type"`
	Name           string         `json:"name"`
	ControlKind    *string        `json:"control_kind,omitempty"`
	WorkTypeID     *int32         `json:"work_type_id,omitempty"`
	WorkTypeName   *string        `json:"work_type_name,omitempty"`
	WorkTypeCode   *string        `json:"work_type_code,omitempty"`
	WorkTypeMeta   map[string]any `json:"work_type_meta,omitempty"`
	InputMapping   []MappingDTO   `json:"input_mapping"`
	CanvasPosition map[string]any `json:"canvas_position,omitempty"`
	InputSchema    map[string]any `json:"input_schema,omitempty"`
	OutputSchema   map[string]any `json:"output_schema,omitempty"`
}

type MappingDTO struct {
	Target string `json:"target"`
	Source string `json:"source"`
}

type GraphDependencyDTO struct {
	StepID          int32  `json:"step_id"`
	DependsOnStepID int32  `json:"depends_on_step_id"`
	Outcome         string `json:"outcome"`
	OutputIndex     int32  `json:"output_index"`
}

type StepRunDetailDTO struct {
	ID           int32          `json:"id"`
	StepID       int32          `json:"step_id"`
	Status       string         `json:"status"`
	Outcome      *string        `json:"outcome,omitempty"`
	InputData    map[string]any `json:"input_data,omitempty"`
	OutputData   map[string]any `json:"output_data,omitempty"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	DurationMs   *int64         `json:"duration_ms,omitempty"`
	ErrorMessage *string        `json:"error_message,omitempty"`
}
