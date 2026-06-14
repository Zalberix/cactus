package runtime

import (
	"context"
	"io"
	"os/exec"
	"time"
)

type TargetKind string

const (
	TargetTask    TargetKind = "task"
	TargetProcess TargetKind = "process"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusSkipped    Status = "skipped"
	StatusCleaning   Status = "cleaning"
	StatusBuilding   Status = "building"
	StatusStarting   Status = "starting"
	StatusReady      Status = "ready"
	StatusRunning    Status = "running"
	StatusRestarting Status = "restarting"
	StatusStopping   Status = "stopping"
	StatusStopped    Status = "stopped"
	StatusFailed     Status = "failed"
)

type EventType string

const (
	EventStatus EventType = "status"
	EventLog    EventType = "log"
	EventExit   EventType = "exit"
)

type Stream string

const (
	StreamSystem Stream = "system"
	StreamStdout Stream = "stdout"
	StreamStderr Stream = "stderr"
)

type LogEntry struct {
	Time     time.Time
	TargetID string
	Stream   Stream
	Line     string
}

type Event struct {
	Type     EventType
	Time     time.Time
	TargetID string
	Status   Status
	Message  string
	Line     *LogEntry
	Err      error
}

type CommandFactory func(ctx context.Context, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error)
type BuildFunc func(ctx context.Context, stdout io.Writer, stderr io.Writer) error
type ReadyProbe func(ctx context.Context) error

type TargetSpec struct {
	ID                  string
	Name                string
	Kind                TargetKind
	DependsOn           []string
	Skip                bool
	Build               BuildFunc
	Command             CommandFactory
	Ready               ReadyProbe
	WatchPaths          []string
	RestartOnFileChange bool
	GracefulTimeout     time.Duration
}

type TargetState struct {
	Spec       TargetSpec
	Status     Status
	PID        int
	LastError  string
	StartedAt  time.Time
	StoppedAt  time.Time
	Restarting bool
}

func StatusEvent(targetID string, status Status, message string) Event {
	return Event{
		Type:     EventStatus,
		Time:     time.Now(),
		TargetID: targetID,
		Status:   status,
		Message:  message,
	}
}

func LogEvent(targetID string, stream Stream, line string) Event {
	entry := LogEntry{
		Time:     time.Now(),
		TargetID: targetID,
		Stream:   stream,
		Line:     line,
	}
	return Event{
		Type:     EventLog,
		Time:     entry.Time,
		TargetID: targetID,
		Line:     &entry,
	}
}

func ExitEvent(targetID string, err error) Event {
	return Event{
		Type:     EventExit,
		Time:     time.Now(),
		TargetID: targetID,
		Err:      err,
	}
}
