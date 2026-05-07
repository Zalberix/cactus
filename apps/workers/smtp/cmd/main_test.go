package main

import "testing"

func TestSMTPWorkerSourceIncludesVariantAndWorkerID(t *testing.T) {
	got := smtpWorkerSource("auth", 3)
	want := "worker/smtp-auth-3"
	if got != want {
		t.Fatalf("smtpWorkerSource() = %q, want %q", got, want)
	}
}

func TestSMTPWorkerNameUsesOverrideWhenProvided(t *testing.T) {
	got := smtpWorkerName("basic", "smtp-basic-2")
	want := "smtp-basic-2"
	if got != want {
		t.Fatalf("smtpWorkerName() = %q, want %q", got, want)
	}
}

func TestSMTPWorkerNameFallsBackToVariantName(t *testing.T) {
	got := smtpWorkerName("basic", "")
	want := "smtp-basic-worker"
	if got != want {
		t.Fatalf("smtpWorkerName() = %q, want %q", got, want)
	}
}
