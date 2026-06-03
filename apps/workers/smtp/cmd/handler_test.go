package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/zalberix/cactus/libs/worker"
)

type recordingSender struct {
	messages []emailMessage
	settings smtpSettings
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
	variants := smtpVariants(nil)

	basic, err := worker.SelectVariant(variants, "basic")
	if err != nil {
		t.Fatalf("select basic: %v", err)
	}
	auth, err := worker.SelectVariant(variants, "auth")
	if err != nil {
		t.Fatalf("select auth: %v", err)
	}

	if basic.Manifest.Kind != "smtp-basic" {
		t.Fatalf("expected basic kind smtp-basic, got %q", basic.Manifest.Kind)
	}
	if auth.Manifest.Kind != "smtp-auth" {
		t.Fatalf("expected auth kind smtp-auth, got %q", auth.Manifest.Kind)
	}
	if basic.Manifest.Type != auth.Manifest.Type || basic.Manifest.Type != "email" {
		t.Fatalf("expected both variants to share email type, got %q and %q", basic.Manifest.Type, auth.Manifest.Type)
	}

	var authInput map[string]any
	if err := json.Unmarshal(auth.Manifest.InputSchema, &authInput); err != nil {
		t.Fatalf("unmarshal auth input schema: %v", err)
	}
	props := authInput["properties"].(map[string]any)
	if _, ok := props["cc"]; !ok {
		t.Fatalf("expected auth input schema to contain cc, got %#v", props)
	}
}

func TestRichSMTPHandlerUsesCCRecipients(t *testing.T) {
	sender := &recordingSender{}
	variants := smtpVariants(func(settings smtpSettings) emailSender {
		sender.settings = settings
		return sender
	})
	auth, err := worker.SelectVariant(variants, "auth")
	if err != nil {
		t.Fatalf("select auth: %v", err)
	}

	result, err := auth.Handler.Handle(context.Background(), worker.TaskMessage{
		Settings: map[string]any{
			"host": "smtp.from.nats",
			"port": float64(2525),
			"from": "noreply@example.test",
			"auth": "none",
			"tls":  "none",
		},
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
	if sender.messages[0].From != "noreply@example.test" {
		t.Fatalf("expected from from task settings, got %q", sender.messages[0].From)
	}
}

func TestSMTPHandlerUsesTaskSettingsFromNATS(t *testing.T) {
	recorder := &recordingSender{}
	handler := SMTPHandler{
		senderFactory: func(settings smtpSettings) emailSender {
			recorder.settings = settings
			return recorder
		},
	}

	_, err := handler.Handle(context.Background(), worker.TaskMessage{
		Settings: map[string]any{
			"host": "smtp.from.nats",
			"port": float64(2525),
			"from": "noreply@example.test",
			"auth": "none",
			"tls":  "none",
		},
		Input: map[string]any{
			"to":      "a@example.test",
			"subject": "Subject",
			"body":    "Body",
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if recorder.settings.host != "smtp.from.nats" {
		t.Fatalf("expected host from task settings, got %q", recorder.settings.host)
	}
	if recorder.settings.port != 2525 {
		t.Fatalf("expected port 2525, got %d", recorder.settings.port)
	}
	if len(recorder.messages) != 1 {
		t.Fatalf("expected one sent message, got %d", len(recorder.messages))
	}
	if recorder.messages[0].From != "noreply@example.test" {
		t.Fatalf("expected from from task settings, got %q", recorder.messages[0].From)
	}
}

func TestSMTPSendTimeoutDoesNotChangeRegisteredSettingsSchema(t *testing.T) {
	var basicSettings map[string]any
	if err := json.Unmarshal(basicSMTPManifest().SettingsSchema, &basicSettings); err != nil {
		t.Fatalf("unmarshal basic settings schema: %v", err)
	}
	basicProps := basicSettings["properties"].(map[string]any)
	if _, ok := basicProps["send_timeout_seconds"]; ok {
		t.Fatalf("basic settings schema must remain backward-compatible without send_timeout_seconds: %#v", basicProps)
	}

	var authSettings map[string]any
	if err := json.Unmarshal(authSMTPManifest().SettingsSchema, &authSettings); err != nil {
		t.Fatalf("unmarshal auth settings schema: %v", err)
	}
	authProps := authSettings["properties"].(map[string]any)
	if _, ok := authProps["send_timeout_seconds"]; ok {
		t.Fatalf("auth settings schema must remain backward-compatible without send_timeout_seconds: %#v", authProps)
	}
}

func TestSMTPSendTimeoutDefaults(t *testing.T) {
	got := smtpSendTimeout(smtpSettings{})

	if got != defaultSMTPSendTimeout {
		t.Fatalf("timeout = %s, want %s", got, defaultSMTPSendTimeout)
	}
}

func TestSMTPSendTimeoutFromSettings(t *testing.T) {
	got := smtpSendTimeout(smtpSettings{sendTimeoutSec: 3})

	if got != 3*time.Second {
		t.Fatalf("timeout = %s, want 3s", got)
	}
}

func TestSMTPSettingsFromTaskParsesSendTimeout(t *testing.T) {
	settings, err := smtpSettingsFromTask(map[string]any{
		"host":                 "smtp.from.nats",
		"port":                 float64(2525),
		"from":                 "noreply@example.test",
		"send_timeout_seconds": float64(4),
	})
	if err != nil {
		t.Fatalf("smtpSettingsFromTask returned error: %v", err)
	}

	if settings.sendTimeoutSec != 4 {
		t.Fatalf("send timeout seconds = %d, want 4", settings.sendTimeoutSec)
	}
}

func TestSMTPVariantsRejectUnknownVariant(t *testing.T) {
	_, err := worker.SelectVariant(smtpVariants(nil), "missing")
	if err == nil {
		t.Fatal("expected unknown variant error")
	}
}
