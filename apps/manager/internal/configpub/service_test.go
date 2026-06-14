package configpub

import (
	"context"
	"testing"
)

type fakeBus struct {
	subject string
	data    any
}

func (b *fakeBus) PublishJS(_ context.Context, subject string, data any) error {
	b.subject = subject
	b.data = data
	return nil
}

func TestPublishRevisionPublishesConfigPayload(t *testing.T) {
	bus := &fakeBus{}
	svc := New(nil, bus)

	err := svc.PublishRevision(context.Background(), 12, 3, 7, 73, []byte(`{"host":"smtp"}`))
	if err != nil {
		t.Fatalf("PublishRevision error: %v", err)
	}
	if bus.subject != "config.org.12.work_type.3.revision.73" {
		t.Fatalf("subject mismatch: %s", bus.subject)
	}
	payload, ok := bus.data.(Payload)
	if !ok {
		t.Fatalf("expected Payload, got %T", bus.data)
	}
	if payload.OrganizationID != 12 || payload.WorkTypeID != 3 || payload.SchemaID != 7 || payload.RevisionID != 73 {
		t.Fatalf("payload scope mismatch: %#v", payload)
	}
	if payload.ConfigHash != "sha256:19405e61143fe0e527de5831a40124ca6382365defc4a8acaded46882f7211ef" {
		t.Fatalf("hash mismatch: %s", payload.ConfigHash)
	}
	if string(payload.SettingsData) != `{"host":"smtp"}` {
		t.Fatalf("settings mismatch: %s", payload.SettingsData)
	}
}
