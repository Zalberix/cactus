package temporal

import (
	"encoding/json"
	"testing"
)

func TestWorkflowEventSerializesRuntimeDetails(t *testing.T) {
	event := WorkflowEvent{
		Type:        "step_update",
		StepID:      12,
		RunStepID:   34,
		StepType:    "task",
		Status:      StepStatusCompleted,
		InputData:   map[string]any{"email": "ada@example.com"},
		OutputData:  map[string]any{"message_id": "abc"},
		StartedAt:   "2026-05-18T08:00:00Z",
		CompletedAt: "2026-05-18T08:00:01Z",
		Error:       "boom",
		Timestamp:   "2026-05-18T08:00:02Z",
	}

	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal workflow event: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal workflow event: %v", err)
	}

	if got["run_step_id"] != float64(34) {
		t.Fatalf("run_step_id = %v", got["run_step_id"])
	}
	if got["started_at"] != "2026-05-18T08:00:00Z" {
		t.Fatalf("started_at = %v", got["started_at"])
	}
	if got["completed_at"] != "2026-05-18T08:00:01Z" {
		t.Fatalf("completed_at = %v", got["completed_at"])
	}

	input, ok := got["input_data"].(map[string]any)
	if !ok || input["email"] != "ada@example.com" {
		t.Fatalf("input_data = %#v", got["input_data"])
	}
	output, ok := got["output_data"].(map[string]any)
	if !ok || output["message_id"] != "abc" {
		t.Fatalf("output_data = %#v", got["output_data"])
	}
}
