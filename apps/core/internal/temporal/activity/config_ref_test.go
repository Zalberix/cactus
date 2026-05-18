package activity

import (
	"testing"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

func TestBuildTaskDispatchUsesOrgScopedSubjectsAndConfigRef(t *testing.T) {
	input := temporaltypes.RunTaskStepInput{
		WorkflowRunID: 1001,
		Step: temporaltypes.StepDef{
			ID:                       55,
			OrganizationID:           12,
			WorkTypeID:               3,
			WorkerSettingsSchemaID:   7,
			WorkerSettingsRevisionID: 73,
		},
	}

	subject, task := buildTaskDispatch(input, 2, map[string]any{"to": "person@example.com"}, []byte(`{"host":"smtp"}`))

	if subject != "task.org.12.work_type.3.schema.7.revision.73.1001" {
		t.Fatalf("subject mismatch: %s", subject)
	}
	if task.ReplyTo != "result.org.12.1001.55" {
		t.Fatalf("reply subject mismatch: %s", task.ReplyTo)
	}
	if task.ConfigRef.ConfigSubject != "config.org.12.work_type.3.revision.73" {
		t.Fatalf("config subject mismatch: %s", task.ConfigRef.ConfigSubject)
	}
	if task.ConfigRef.ConfigHash != "sha256:19405e61143fe0e527de5831a40124ca6382365defc4a8acaded46882f7211ef" {
		t.Fatalf("config hash mismatch: %s", task.ConfigRef.ConfigHash)
	}
	if task.ConfigRef.OrganizationID != 12 || task.ConfigRef.WorkTypeID != 3 || task.ConfigRef.SchemaID != 7 || task.ConfigRef.RevisionID != 73 {
		t.Fatalf("config ref mismatch: %#v", task.ConfigRef)
	}
}
