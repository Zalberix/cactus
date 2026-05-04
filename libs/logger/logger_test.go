package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestTextHandlerWritesSingleLineWithSourceAndAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := newTextHandler(&buf, slog.LevelDebug, "core", false)

	record := slog.NewRecord(time.Date(2026, 5, 3, 12, 34, 56, 789000000, time.UTC), slog.LevelInfo, "server started", 0)
	record.AddAttrs(
		slog.String("addr", ":3009"),
		slog.Int("workers", 2),
	)

	if err := h.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if strings.Count(got, "\n") != 0 {
		t.Fatalf("expected one log line, got %q", got)
	}
	if !strings.Contains(got, "[12:34:56.789] INFO [core] server started") {
		t.Fatalf("expected source and message in line, got %q", got)
	}
	if !strings.Contains(got, "addr=:3009") || !strings.Contains(got, "workers=2") {
		t.Fatalf("expected attrs in line, got %q", got)
	}
}

func TestTextHandlerEscapesMultilineAndTruncatesLongValues(t *testing.T) {
	var buf bytes.Buffer
	h := newTextHandler(&buf, slog.LevelDebug, "worker/smtp-uuid", false)

	record := slog.NewRecord(time.Date(2026, 5, 3, 12, 34, 56, 0, time.UTC), slog.LevelWarn, "config reloaded", 0)
	record.AddAttrs(
		slog.String("config", "line one\nline two"),
		slog.String("payload", strings.Repeat("x", maxFieldValueLen+20)),
	)

	if err := h.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if strings.Count(got, "\n") != 0 {
		t.Fatalf("expected one log line, got %q", got)
	}
	if strings.Contains(got, "line one\nline two") {
		t.Fatalf("expected multiline value to be escaped, got %q", got)
	}
	if !strings.Contains(got, `config="line one\nline two"`) {
		t.Fatalf("expected quoted escaped multiline value, got %q", got)
	}
	if !strings.Contains(got, "payload=") || !strings.Contains(got, "truncated=true") {
		t.Fatalf("expected long value truncation marker, got %q", got)
	}
}

func TestTextHandlerWithAttrsIncludesInheritedAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := newTextHandler(&buf, slog.LevelDebug, "core", false).WithAttrs([]slog.Attr{
		slog.String("component", "http"),
	})

	record := slog.NewRecord(time.Date(2026, 5, 3, 12, 34, 56, 0, time.UTC), slog.LevelError, "failed", 0)
	record.AddAttrs(slog.String("error", "boom"))

	if err := h.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if !strings.Contains(got, "component=http") || !strings.Contains(got, "error=boom") {
		t.Fatalf("expected inherited and record attrs, got %q", got)
	}
}

func TestWorkerSourceUsesWorkerID(t *testing.T) {
	got := WorkerSource("smtp", 42)
	want := "worker/smtp-42"
	if got != want {
		t.Fatalf("WorkerSource() = %q, want %q", got, want)
	}
}

func TestWorkerSourceWithoutWorkerID(t *testing.T) {
	got := WorkerSource("smtp", 0)
	want := "worker/smtp"
	if got != want {
		t.Fatalf("WorkerSource() = %q, want %q", got, want)
	}
}
