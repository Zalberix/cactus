package wshub

import (
	"testing"

	"github.com/zalberix/cactus/apps/core/internal/domain/message"
	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

func TestBuildSnapshotUsesMessageDetail(t *testing.T) {
	detail := &message.MessageDetailResponse{
		MessageID: 100,
		WorkflowRun: &message.WorkflowRunStatus{
			ID:     91,
			Status: temporaltypes.RunStatusFailed,
		},
	}

	snapshot := (&Hub{}).buildSnapshot(detail)
	if snapshot.Type != "snapshot" {
		t.Fatalf("type = %q", snapshot.Type)
	}
	if snapshot.Detail != detail {
		t.Fatalf("snapshot detail was not preserved")
	}
}

func TestTerminalFailedEventIncludesRunError(t *testing.T) {
	errMsg := "smtp failed"
	detail := &message.MessageDetailResponse{
		WorkflowRun: &message.WorkflowRunStatus{
			Status:       temporaltypes.RunStatusFailed,
			ErrorMessage: &errMsg,
		},
	}

	event := terminalEventForStatus(temporaltypes.RunStatusFailed, detail, "2026-05-18T08:00:00Z")
	failed, ok := event.(WorkflowFailedEvent)
	if !ok {
		t.Fatalf("event type = %T", event)
	}
	if failed.Error != errMsg {
		t.Fatalf("error = %q", failed.Error)
	}
}
