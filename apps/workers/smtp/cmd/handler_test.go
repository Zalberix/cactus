package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/zalberix/cactus/apps/workers/smtp/config"
	"github.com/zalberix/cactus/libs/worker"
)

type recordingSender struct {
	messages []emailMessage
}

func (s *recordingSender) Send(_ context.Context, msg emailMessage) (emailSendResult, error) {
	s.messages = append(s.messages, msg)
	return emailSendResult{
		MessageID:       "msg-1",
		SentAt:          "2026-05-06T00:00:00Z",
		RecipientsCount: len(msg.To) + len(msg.CC),
	}, nil
}

func TestSMTPVariantsExposeDifferentManifests(t *testing.T) {
	variants := smtpVariants(config.Config{}, &recordingSender{})

	basic, err := worker.SelectVariant(variants, "basic")
	if err != nil {
		t.Fatalf("select basic: %v", err)
	}
	rich, err := worker.SelectVariant(variants, "rich")
	if err != nil {
		t.Fatalf("select rich: %v", err)
	}

	if basic.Manifest.Kind != "smtp-basic" {
		t.Fatalf("expected basic kind smtp-basic, got %q", basic.Manifest.Kind)
	}
	if rich.Manifest.Kind != "smtp-rich" {
		t.Fatalf("expected rich kind smtp-rich, got %q", rich.Manifest.Kind)
	}
	if basic.Manifest.Type != rich.Manifest.Type || basic.Manifest.Type != "email" {
		t.Fatalf("expected both variants to share email type, got %q and %q", basic.Manifest.Type, rich.Manifest.Type)
	}

	var richInput map[string]any
	if err := json.Unmarshal(rich.Manifest.InputSchema, &richInput); err != nil {
		t.Fatalf("unmarshal rich input schema: %v", err)
	}
	props := richInput["properties"].(map[string]any)
	if _, ok := props["cc"]; !ok {
		t.Fatalf("expected rich input schema to contain cc, got %#v", props)
	}
}

func TestRichSMTPHandlerUsesCCRecipients(t *testing.T) {
	sender := &recordingSender{}
	variants := smtpVariants(config.Config{}, sender)
	rich, err := worker.SelectVariant(variants, "rich")
	if err != nil {
		t.Fatalf("select rich: %v", err)
	}

	result, err := rich.Handler.Handle(context.Background(), worker.TaskMessage{
		Input: map[string]any{
			"to":      "a@example.test",
			"cc":      []any{"b@example.test", "c@example.test"},
			"subject": "Subject",
			"body":    "Body",
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success result: %#v", result)
	}
	if result.Output["recipients_count"] != 3 {
		t.Fatalf("expected 3 recipients, got %#v", result.Output["recipients_count"])
	}
	if len(sender.messages) != 1 {
		t.Fatalf("expected one sent message, got %d", len(sender.messages))
	}
	if len(sender.messages[0].CC) != 2 {
		t.Fatalf("expected two cc recipients, got %#v", sender.messages[0].CC)
	}
}

func TestSMTPVariantsRejectUnknownVariant(t *testing.T) {
	_, err := worker.SelectVariant(smtpVariants(config.Config{}, &recordingSender{}), "missing")
	if err == nil {
		t.Fatal("expected unknown variant error")
	}
}
