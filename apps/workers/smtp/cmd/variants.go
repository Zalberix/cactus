package main

import (
	"context"
	"fmt"
	"time"

	mail "github.com/wneessen/go-mail"

	"github.com/zalberix/cactus/apps/workers/smtp/config"
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
		return emailSendResult{}, fmt.Errorf("create smtp client: %w", err)
	}
	if err := client.DialAndSend(m); err != nil {
		return emailSendResult{}, fmt.Errorf("send email: %w", err)
	}

	return emailSendResult{
		MessageID:       m.GetMessageID(),
		SentAt:          time.Now().Format(time.RFC3339),
		RecipientsCount: len(msg.To) + len(msg.CC),
	}, nil
}

type SMTPHandler struct {
	from        string
	sender      emailSender
	includeCC   bool
	requireBody bool
}

func (h SMTPHandler) Handle(ctx context.Context, task worker.TaskMessage) (worker.Result, error) {
	to, _ := task.Input["to"].(string)
	subject, _ := task.Input["subject"].(string)
	body, _ := task.Input["body"].(string)
	if to == "" || subject == "" {
		return worker.Result{}, fmt.Errorf("missing required fields: to=%q, subject=%q", to, subject)
	}
	if h.requireBody && body == "" {
		return worker.Result{}, fmt.Errorf("missing required field: body")
	}

	msg := emailMessage{
		From:    h.from,
		To:      []string{to},
		Subject: subject,
		Body:    body,
	}
	if h.includeCC {
		msg.CC = stringsFromInput(task.Input["cc"])
	}

	sendResult, err := h.sender.Send(ctx, msg)
	if err != nil {
		return worker.Result{}, err
	}

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

func smtpVariants(cfg config.Config, sender emailSender) map[string]worker.Variant {
	if sender == nil {
		sender = smtpSender{
			host: cfg.SMTP.Host,
			port: cfg.SMTP.Port,
			auth: cfg.SMTP.Auth,
			tls:  cfg.SMTP.TLS,
		}
	}

	return map[string]worker.Variant{
		"basic": {
			Name:     "basic",
			Manifest: basicSMTPManifest(),
			Handler:  SMTPHandler{from: cfg.SMTP.From, sender: sender},
		},
		"auth": {
			Name:     "auth",
			Manifest: authSMTPManifest(),
			Handler:  SMTPHandler{from: cfg.SMTP.From, sender: sender, includeCC: true, requireBody: true},
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
