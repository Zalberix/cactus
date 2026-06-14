package runtime

import "testing"

func TestEventConstructorsPopulateFields(t *testing.T) {
	status := StatusEvent("manager", StatusStarting, "starting")
	if status.Type != EventStatus || status.TargetID != "manager" || status.Status != StatusStarting || status.Message != "starting" {
		t.Fatalf("unexpected status event: %#v", status)
	}
	if status.Time.IsZero() {
		t.Fatal("status event time was not set")
	}

	log := LogEvent("manager", StreamStdout, "hello")
	if log.Type != EventLog || log.TargetID != "manager" || log.Line == nil {
		t.Fatalf("unexpected log event: %#v", log)
	}
	if log.Line.TargetID != "manager" || log.Line.Stream != StreamStdout || log.Line.Line != "hello" {
		t.Fatalf("unexpected log entry: %#v", log.Line)
	}
	if log.Time.IsZero() || log.Line.Time.IsZero() {
		t.Fatal("log event time was not set")
	}

	err := assertErr("boom")
	exit := ExitEvent("manager", err)
	if exit.Type != EventExit || exit.TargetID != "manager" || exit.Err != err {
		t.Fatalf("unexpected exit event: %#v", exit)
	}
	if exit.Time.IsZero() {
		t.Fatal("exit event time was not set")
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
