package message

import "time"

// SendMessageRequest — запрос на отправку сообщения (per D-10).
// JSON body: {"workflow_id": 1, "value": {...}, "external_id": "ext-123"}
type SendMessageRequest struct {
	WorkflowID int32          `json:"workflow_id" binding:"required"`
	Value      map[string]any `json:"value" binding:"required"`
	ExternalID string         `json:"external_id"`
}

// SendMessageResponse — ответ на успешную отправку (per D-14).
type SendMessageResponse struct {
	MessageID     int32  `json:"message_id"`
	WorkflowRunID int32  `json:"workflow_run_id"`
	Status        string `json:"status"` // "running"
}

// MessageStatusResponse — ответ Status API (per D-20, D-21). Envelope format.
type MessageStatusResponse struct {
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

// StepStatusDTO — step status without input/output (per D-21).
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
