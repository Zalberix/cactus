package natssubjects

import "testing"

func TestSubjects(t *testing.T) {
	scope := WorkerScope{OrganizationID: 12, WorkTypeID: 3}

	if got := Task(scope, 7, 73, 1001); got != "task.org.12.work_type.3.schema.7.revision.73.1001" {
		t.Fatalf("Task subject mismatch: %s", got)
	}
	if got := TaskFilter(scope); got != "task.org.12.work_type.3.>" {
		t.Fatalf("Task filter subject mismatch: %s", got)
	}
	if got := Result(12, 1001, 55); got != "result.org.12.1001.55" {
		t.Fatalf("Result subject mismatch: %s", got)
	}
	if got := Config(12, 3, 73); got != "config.org.12.work_type.3.revision.73" {
		t.Fatalf("Config subject mismatch: %s", got)
	}
	if got := ConfigRequest(12, 3, 73); got != "config.request.org.12.work_type.3.revision.73" {
		t.Fatalf("Config request subject mismatch: %s", got)
	}
}
