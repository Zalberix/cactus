package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	mail "github.com/wneessen/go-mail"

	"github.com/zalberix/cactus/libs/worker"
)

type emailMessage struct {
	From    string
	To      []string
	CC      []string
	Subject string
	Body    string
}

type emailSendResult struct {
	MessageID       string
	SentAt          string
	RecipientsCount int
}

type emailSender interface {
	Send(ctx context.Context, msg emailMessage) (emailSendResult, error)
}

type smtpSettings struct {
	host string
	port int
	from string
	auth string
	tls  string
}

type smtpSender struct {
	host string
	port int
	auth string
	tls  string
}

func (s smtpSender) Send(_ context.Context, msg emailMessage) (emailSendResult, error) {
	m := mail.NewMsg()
	if err := m.From(msg.From); err != nil {
		return emailSendResult{}, fmt.Errorf("set from %q: %w", msg.From, err)
	}
	if err := m.To(msg.To...); err != nil {
		return emailSendResult{}, fmt.Errorf("set to: %w", err)
	}
	if len(msg.CC) > 0 {
		if err := m.Cc(msg.CC...); err != nil {
			return emailSendResult{}, fmt.Errorf("set cc: %w", err)
		}
	}
	m.Subject(msg.Subject)
	m.SetBodyString(mail.TypeTextHTML, msg.Body)

	slog.Info("smtp sender prepared message",
		slog.String("from", msg.From),
		slog.String("to", strings.Join(msg.To, ",")),
		slog.String("subject", msg.Subject),
		slog.Int("recipients_to", len(msg.To)),
		slog.Int("recipients_cc", len(msg.CC)),
	)

	opts := []mail.Option{mail.WithPort(s.port)}
	switch s.auth {
	case "none", "":
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthNoAuth))
	}
	switch s.tls {
	case "none", "":
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	case "tls":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	case "starttls":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(s.host, opts...)
	if err != nil {
		slog.Error("smtp sender create client failed",
			slog.String("host", s.host),
			slog.Int("port", s.port),
			slog.String("error", err.Error()),
		)
		return emailSendResult{}, fmt.Errorf("create smtp client: %w", err)
	}

	if err := client.DialAndSend(m); err != nil {
		slog.Error("smtp sender send failed",
			slog.String("host", s.host),
			slog.Int("port", s.port),
			slog.String("subject", msg.Subject),
			slog.String("error", err.Error()),
		)
		return emailSendResult{}, fmt.Errorf("send email: %w", err)
	}

	slog.Info("smtp sender sent message",
		slog.String("message_id", m.GetMessageID()),
		slog.String("subject", msg.Subject),
		slog.Int("recipients_count", len(msg.To)+len(msg.CC)),
	)

	return emailSendResult{
		MessageID:       m.GetMessageID(),
		SentAt:          time.Now().Format(time.RFC3339),
		RecipientsCount: len(msg.To) + len(msg.CC),
	}, nil
}

type SMTPHandler struct {
	senderFactory func(smtpSettings) emailSender
	includeCC      bool
	requireBody    bool
}

func (h SMTPHandler) Handle(ctx context.Context, task worker.TaskMessage) (worker.Result, error) {
	settings, err := smtpSettingsFromTask(task.Settings)
	if err != nil {
		return worker.Result{}, err
	}
	senderFactory := h.senderFactory
	if senderFactory == nil {
		senderFactory = func(settings smtpSettings) emailSender {
			return smtpSender{
				host: settings.host,
				port: settings.port,
				auth: settings.auth,
				tls:  settings.tls,
			}
		}
	}
	sender := senderFactory(settings)

	to, _ := task.Input["to"].(string)
	subject, _ := task.Input["subject"].(string)
	body, _ := task.Input["body"].(string)

	log := slog.Default()
	cc := []string(nil)
	if h.includeCC {
		cc = stringsFromInput(task.Input["cc"])
	}

	log.Info("SMTP handler received task",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
		slog.String("to", to),
		slog.String("subject", subject),
		slog.Int("cc_count", len(cc)),
		slog.Int("input_len", len(task.Input)),
		slog.String("idempotency_key", task.IdempotencyKey),
		slog.String("reply_to", task.ReplyTo),
	)

	if to == "" || subject == "" {
		log.Warn("SMTP handler missing required fields",
			slog.Int("workflow_run_id", int(task.WorkflowRunID)),
			slog.Int("step_id", int(task.StepID)),
			slog.String("to", to),
			slog.String("subject", subject),
		)
		return worker.Result{}, fmt.Errorf("missing required fields: to=%q, subject=%q", to, subject)
	}
	if h.requireBody && body == "" {
		log.Warn("SMTP handler missing required body",
			slog.Int("workflow_run_id", int(task.WorkflowRunID)),
			slog.Int("step_id", int(task.StepID)),
		)
		return worker.Result{}, fmt.Errorf("missing required field: body")
	}

	msg := emailMessage{
		From:    settings.from,
		To:      []string{to},
		Subject: subject,
		Body:    body,
		CC:      cc,
	}

	log.Info("SMTP handler sending message",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
		slog.Int("recipients", len(msg.To)+len(msg.CC)),
		slog.String("reply_to", task.ReplyTo),
	)

	sendResult, err := sender.Send(ctx, msg)
	if err != nil {
		log.Error("SMTP handler send error",
			slog.Int("workflow_run_id", int(task.WorkflowRunID)),
			slog.Int("step_id", int(task.StepID)),
			slog.Int("attempt", int(task.Attempt)),
			slog.String("to", to),
			slog.String("subject", subject),
			slog.String("error", err.Error()),
		)
		return worker.Result{}, err
	}

	log.Info("SMTP handler send success",
		slog.Int("workflow_run_id", int(task.WorkflowRunID)),
		slog.Int("step_id", int(task.StepID)),
		slog.Int("attempt", int(task.Attempt)),
		slog.String("smtp_message_id", sendResult.MessageID),
		slog.String("sent_at", sendResult.SentAt),
		slog.Int("recipients_count", sendResult.RecipientsCount),
	)

	return worker.Result{
		Success: true,
		Output: map[string]any{
			"message_id":       sendResult.MessageID,
			"sent_at":          sendResult.SentAt,
			"recipients_count": sendResult.RecipientsCount,
		},
	}, nil
}

func stringsFromInput(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && text != "" {
				result = append(result, text)
			}
		}
		return result
	case string:
		if typed == "" {
			return nil
		}
		return []string{typed}
	default:
		return nil
	}
}

func smtpSettingsFromTask(settings map[string]any) (smtpSettings, error) {
	host, ok := stringSetting(settings, "host")
	if !ok {
		return smtpSettings{}, fmt.Errorf("missing required smtp setting: host")
	}
	port, ok := intSetting(settings, "port")
	if !ok || port <= 0 {
		return smtpSettings{}, fmt.Errorf("missing required smtp setting: port")
	}
	from, ok := stringSetting(settings, "from")
	if !ok {
		return smtpSettings{}, fmt.Errorf("missing required smtp setting: from")
	}
	auth, _ := stringSetting(settings, "auth")
	if auth == "" {
		auth = "none"
	}
	tlsMode, _ := stringSetting(settings, "tls")
	if tlsMode == "" {
		tlsMode = "none"
	}
	return smtpSettings{
		host: host,
		port: port,
		from: from,
		auth: auth,
		tls:  tlsMode,
	}, nil
}

func stringSetting(settings map[string]any, key string) (string, bool) {
	value, ok := settings[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.TrimSpace(text)
	return text, text != ""
}

func intSetting(settings map[string]any, key string) (int, bool) {
	value, ok := settings[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		if typed > math.MaxInt || typed < math.MinInt {
			return 0, false
		}
		return int(typed), true
	case float64:
		if typed != math.Trunc(typed) || typed > float64(math.MaxInt) || typed < float64(math.MinInt) {
			return 0, false
		}
		return int(typed), true
	case float32:
		f64 := float64(typed)
		if f64 != math.Trunc(f64) || f64 > float64(math.MaxInt) || f64 < float64(math.MinInt) {
			return 0, false
		}
		return int(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil || parsed > math.MaxInt || parsed < math.MinInt {
			return 0, false
		}
		return int(parsed), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func smtpVariants(senderFactory func(smtpSettings) emailSender) map[string]worker.Variant {
	return map[string]worker.Variant{
		"basic": {
			Name:     "basic",
			Manifest: basicSMTPManifest(),
			Handler:  SMTPHandler{senderFactory: senderFactory},
		},
		"auth": {
			Name:     "auth",
			Manifest: authSMTPManifest(),
			Handler:  SMTPHandler{senderFactory: senderFactory, includeCC: true, requireBody: true},
		},
	}
}

func basicSMTPManifest() worker.ManifestSpec {
	return worker.Manifest().
		Kind("smtp-basic", "SMTP Basic Email").
		Type("email", "Email Delivery").
		SettingsSchema(func(sb *worker.SchemaBuilder) {
			sb.String("host").Required()
			sb.Integer("port").Required()
			sb.String("from").Required()
		}).
		InputSchema(func(sb *worker.SchemaBuilder) {
			sb.String("to").Required()
			sb.String("subject").Required()
			sb.String("body")
		}).
		OutputSchema(func(sb *worker.SchemaBuilder) {
			sb.String("message_id")
			sb.String("sent_at")
			sb.Integer("recipients_count")
		}).
		Build()
}

func authSMTPManifest() worker.ManifestSpec {
	return worker.Manifest().
		Kind("smtp-auth", "SMTP Auth Email").
		Type("email", "Email Delivery").
		SettingsSchema(func(sb *worker.SchemaBuilder) {
			sb.String("host").Required()
			sb.Integer("port").Required()
			sb.String("from").Required()
			sb.String("auth").Enum("none", "plain", "login")
			sb.String("tls").Enum("none", "tls", "starttls")
		}).
		InputSchema(func(sb *worker.SchemaBuilder) {
			sb.String("to").Required()
			sb.String("cc")
			sb.String("subject").Required()
			sb.String("body").Required()
		}).
		OutputSchema(func(sb *worker.SchemaBuilder) {
			sb.String("message_id")
			sb.String("sent_at")
			sb.Integer("recipients_count")
		}).
		Build()
}
