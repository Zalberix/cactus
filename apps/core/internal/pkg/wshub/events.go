package wshub

// AuthMessage -- client sends this as first WS message (per D-13).
// Token field carries either a JWT string OR a system token in
// "public_token:private_token" format (colon-separated).
type AuthMessage struct {
	Type  string `json:"type"`  // must be "auth"
	Token string `json:"token"` // JWT or "public_token:private_token"
}

// AuthOK -- server response on successful auth.
type AuthOK struct {
	Type string `json:"type"` // "auth_ok"
}

// SnapshotEvent -- sent once after auth_ok (per D-14).
// Contains current state from DB.
type SnapshotEvent struct {
	Type           string         `json:"type"`            // "snapshot"
	WorkflowStatus string         `json:"workflow_status"` // "running", "completed", "failed", etc.
	Steps          []SnapshotStep `json:"steps"`
}

// SnapshotStep -- step in snapshot.
type SnapshotStep struct {
	StepID      int32   `json:"step_id"`
	StepType    string  `json:"step_type"`
	Status      string  `json:"status"`
	StartedAt   *string `json:"started_at"`   // RFC3339 or null
	CompletedAt *string `json:"completed_at"` // RFC3339 or null
}

// StepUpdateEvent -- delta event for step status change (per D-15).
// Same shape as WorkflowEvent from NATS -- Hub forwards without transformation.
type StepUpdateEvent struct {
	Type      string `json:"type"` // "step_update"
	StepID    int32  `json:"step_id"`
	StepType  string `json:"step_type"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"` // RFC3339
}

// WorkflowDoneEvent -- terminal event (per D-15, D-18).
type WorkflowDoneEvent struct {
	Type      string `json:"type"` // "workflow_done"
	Timestamp string `json:"timestamp"`
}

// WorkflowFailedEvent -- terminal event (per D-15, D-18).
type WorkflowFailedEvent struct {
	Type      string `json:"type"` // "workflow_failed"
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}
