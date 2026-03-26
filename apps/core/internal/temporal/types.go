package temporal

import "time"

// DAGInput — input для DAGExecutorWorkflow (передаётся в Temporal).
type DAGInput struct {
	WorkflowRunID int32     `json:"workflow_run_id"`
	MessageID     int32     `json:"message_id"`
	MessageValue  []byte    `json:"message_value"` // JSON payload от клиента
	Steps         []StepDef `json:"steps"`
	Deps          []DepDef  `json:"deps"`
}

// StepDef — определение шага для workflow executor.
type StepDef struct {
	ID                       int32          `json:"id"`        // workflow_step.id
	StepType                 string         `json:"step_type"` // "task" или "control"
	WorkTypeID               int32          `json:"work_type_id,omitempty"`
	WorkerSettingsRevisionID int32          `json:"worker_settings_revision_id,omitempty"`
	InputMapping             []MappingEntry `json:"input_mapping,omitempty"`
	Timeout                  time.Duration  `json:"timeout,omitempty"` // per D-05: timeout из ревизии настроек воркера
}

// MappingEntry — запись input mapping.
type MappingEntry struct {
	Target string `json:"target"`
	Source string `json:"source"`
}

// DepDef — зависимость между шагами.
type DepDef struct {
	StepID          int32  `json:"step_id"`
	DependsOnStepID int32  `json:"depends_on_step_id"`
	Outcome         string `json:"outcome"`
}

// StepResult — результат выполнения одного шага.
type StepResult struct {
	StepID   int32          `json:"step_id"`
	Success  bool           `json:"success"`
	Output   map[string]any `json:"output,omitempty"`
	Error    string         `json:"error,omitempty"`
	WorkerID int32          `json:"worker_id,omitempty"`
	Outcome  string         `json:"outcome"` // "success" для task шагов
}

// TaskMessage — сообщение воркеру в NATS (per D-06).
type TaskMessage struct {
	WorkflowRunID  int32          `json:"workflow_run_id"`
	StepID         int32          `json:"step_id"`
	Attempt        int32          `json:"attempt"`
	ReplyTo        string         `json:"reply_to"`
	Input          map[string]any `json:"input"`
	IdempotencyKey string         `json:"idempotency_key"` // D-04: runID.stepID.attempt
}

// WorkerResult — результат от воркера (per D-08, расширенный worker_id).
type WorkerResult struct {
	Success  bool           `json:"success"`
	Output   map[string]any `json:"output,omitempty"`
	Error    string         `json:"error,omitempty"`
	WorkerID int32          `json:"worker_id,omitempty"`
}

// RunTaskStepInput — input для RunTaskStep activity.
type RunTaskStepInput struct {
	WorkflowRunID int32              `json:"workflow_run_id"`
	Step          StepDef            `json:"step"`
	Attempt       int32              `json:"attempt"`
	MessageValue  []byte             `json:"message_value"`
	StepOutputs   map[int32]map[string]any `json:"step_outputs"`
}

// RecordStepInput — input для RecordStep activity.
type RecordStepInput struct {
	WorkflowRunID int32          `json:"workflow_run_id"`
	StepID        int32          `json:"step_id"` // workflow_step.id
	Status        string         `json:"status"`  // pending, running, completed, failed, skipped
	InputData     map[string]any `json:"input_data,omitempty"`
	OutputData    map[string]any `json:"output_data,omitempty"`
	ErrorMessage  string         `json:"error_message,omitempty"`
	WorkerID      int32          `json:"worker_id,omitempty"`
	Outcome       string         `json:"outcome,omitempty"`
	AttemptNumber int32          `json:"attempt_number,omitempty"`
}

// Статусы (per D-19, D-20).
const (
	StepStatusPending   = "pending"
	StepStatusRunning   = "running"
	StepStatusCompleted = "completed"
	StepStatusFailed    = "failed"
	StepStatusSkipped   = "skipped"

	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"

	TaskQueueName = "cactus-core"
)
