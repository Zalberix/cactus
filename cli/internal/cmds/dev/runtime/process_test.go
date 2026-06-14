package runtime

import (
	"context"
	"io"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestManagedProcessEmitsReadyAndLogs(t *testing.T) {
	events := make(chan Event, 20)
	spec := TargetSpec{
		ID:   "test-process",
		Name: "test-process",
		Kind: TargetProcess,
		Command: func(ctx context.Context, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
			if runtime.GOOS == "windows" {
				cmd := exec.CommandContext(ctx, "cmd", "/C", "echo hello")
				cmd.Stdout = stdout
				cmd.Stderr = stderr
				return cmd, nil
			}
			cmd := exec.CommandContext(ctx, "sh", "-c", "echo hello")
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			return cmd, nil
		},
		Ready: ImmediateReady(),
	}

	proc := NewManagedProcess(spec, events)
	if err := proc.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	var sawReady bool
	var sawLog bool
	deadline := time.After(time.Second)
	for !(sawReady && sawLog) {
		select {
		case ev := <-events:
			if ev.Type == EventStatus && ev.Status == StatusReady {
				sawReady = true
			}
			if ev.Type == EventLog && ev.Line != nil && ev.Line.Line == "hello" {
				sawLog = true
			}
		case <-deadline:
			t.Fatalf("did not receive ready and log events")
		}
	}
}
