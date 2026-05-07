package main

import "testing"

func TestSMTPWorkerSourceIncludesVariantAndWorkerID(t *testing.T) {
	got := smtpWorkerSource("auth", 3)
	want := "worker/smtp-auth-3"
	if got != want {
		t.Fatalf("smtpWorkerSource() = %q, want %q", got, want)
	}
}
