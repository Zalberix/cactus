package message

import "time"

// SendMessageRequest — запрос на отправку сообщения
// JSON body: {"workflow_id": 1, "value": {...}, "external_id": "ext-123"}
type SendMessageRequest struct {
	WorkflowID int32          `json:"workflow_id" binding:"required"`
	Value      map[string]any `json:"value" binding:"required"`
	ExternalID string         `json:"external_id"`
}

// SendMessageResponse — ответ на успешную отправку
type SendMessageResponse struct {
	MessageID     int32  `json:"message_id"`
	WorkflowRunID int32  `json:"workflow_run_id"`
	Status        string `json:"status"` // "running"
}

// StatusResponse is the Status API envelope.
type StatusResponse struct {
	MessageID     int32              `json:"message_id"`
	MessageStatus string             `json:"message_status"`
	CreatedAt     time.Time          `json:"created_at"`
	WorkflowRun   *WorkflowRunStatus `json:"workflow_run,omitempty"`
	Steps         []StepStatusDTO    `json:"steps"`
}

// WorkflowRunStatus — nested workflow_run info in status response.
type WorkflowRunStatus struct {
	ID           int32      `json:"id"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	ErrorMessage *string    `json:"error_message"`
}

// StepStatusDTO — step status without input/output.
type StepStatusDTO struct {
	ID           int32      `json:"id"`
	StepID       int32      `json:"step_id"`
	StepType     string     `json:"step_type"`
	Status       string     `json:"status"`
	Outcome      *string    `json:"outcome"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	ErrorMessage *string    `json:"error_message"`
}

// ListItem is a single message in organization listing.
type ListItem struct {
	ID           int32  `json:"id"`
	WorkflowID   int32  `json:"workflow_id"`
	WorkflowName string `json:"workflow_name"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type DetailResponse struct {
	MessageID     int32              `json:"message_id"`
	WorkflowID    int32              `json:"workflow_id"`
	WorkflowName  string             `json:"workflow_name"`
	MessageStatus string             `json:"message_status"`
	MessageValue  map[string]any     `json:"message_value"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	WorkflowRun   *WorkflowRunStatus `json:"workflow_run,omitempty"`
	Graph         GraphDTO           `json:"graph"`
	RunSteps      []StepRunDetailDTO `json:"run_steps"`
}

type GraphDTO struct {
	VersionID    int32                `json:"version_id"`
	Steps        []GraphStepDTO       `json:"steps"`
	Dependencies []GraphDependencyDTO `json:"dependencies"`
}

type GraphStepDTO struct {
	ID             int32          `json:"id"`
	StepType       string         `json:"step_type"`
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
	ErrorMessage *string        `json:"error_message,omitempty"`
}
