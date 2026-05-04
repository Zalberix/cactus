package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	mail "github.com/wneessen/go-mail"

	"github.com/zalberix/cactus/apps/workers/smtp/config"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
	"github.com/zalberix/cactus/libs/worker"
)

func main() {
	workerIDPath := flag.String("worker-id-path", "", "path to worker ID file (.worker_id/{type}/{uuid})")
	flag.Parse()

	cfg := cfgloader.MustLoad[config.Config]("configs/workers/smtp.yaml")
	slog.SetDefault(logger.SetupLogger(cfg.Env, logger.WorkerSource("smtp", 0)))
	log := slog.Default()

	handler := &SMTPHandler{
		smtpHost: cfg.SMTP.Host,
		smtpPort: cfg.SMTP.Port,
		from:     cfg.SMTP.From,
		auth:     cfg.SMTP.Auth,
		tls:      cfg.SMTP.TLS,
	}

	w := worker.New(worker.Config{
		NatsURL:        cfg.NatsURL,
		ManagerURL:     cfg.ManagerURL,
		BootstrapToken: cfg.BootstrapToken,
		WorkTypeID:     cfg.WorkTypeID,
		RevisionID:     cfg.RevisionID,
		WorkerIDPath:   *workerIDPath,
		WorkerName:     "smtp-worker",
		OnWorkerID: func(workerID int32) *slog.Logger {
			return logger.SetupLogger(cfg.Env, logger.WorkerSource("smtp", workerID))
		},
		Manifest: worker.Manifest{
			Kind:        "smtp",
			NameKind:    "SMTP Email",
			Type:        "email",
			NameType:    "Email Delivery",
			InputSchema: []byte(`{"to":"string","subject":"string","body":"string"}`),
		},
	}, handler, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("SMTP worker starting")

	if err := w.Run(ctx); err != nil {
		slog.Error("worker stopped", slog.String("error", err.Error()))
		cancel()
		return
	}

	slog.Info("SMTP worker stopped")
}

// SMTPHandler реализует worker.TaskHandler для отправки email через go-mail.
type SMTPHandler struct {
	smtpHost string
	smtpPort int
	from     string
	auth     string
	tls      string
}

// Handle обрабатывает задачу отправки email
// Извлекает to, subject, body из task.Input, отправляет через SMTP.
func (h *SMTPHandler) Handle(_ context.Context, task worker.TaskMessage) (worker.Result, error) {
	to, _ := task.Input["to"].(string)
	subject, _ := task.Input["subject"].(string)
	body, _ := task.Input["body"].(string)

	if to == "" || subject == "" {
		return worker.Result{}, fmt.Errorf("missing required fields: to=%q, subject=%q", to, subject)
	}

	m := mail.NewMsg()
	if err := m.From(h.from); err != nil {
		return worker.Result{}, fmt.Errorf("set from %q: %w", h.from, err)
	}
	if err := m.To(to); err != nil {
		return worker.Result{}, fmt.Errorf("set to %q: %w", to, err)
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, body)

	// SMTP client options (per D-11: MailHog = no auth, no TLS)
	opts := []mail.Option{
		mail.WithPort(h.smtpPort),
	}

	switch h.auth {
	case "none", "":
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthNoAuth))
	}

	switch h.tls {
	case "none", "":
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	case "tls":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	case "starttls":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	c, err := mail.NewClient(h.smtpHost, opts...)
	if err != nil {
		return worker.Result{}, fmt.Errorf("create smtp client: %w", err)
	}

	if err := c.DialAndSend(m); err != nil {
		return worker.Result{}, fmt.Errorf("send email: %w", err)
	}

	// Расширенный output (per D-10)
	return worker.Result{
		Success: true,
		Output: map[string]any{
			"message_id":       m.GetMessageID(),
			"sent_at":          time.Now().Format(time.RFC3339),
			"recipients_count": 1,
		},
	}, nil
}
