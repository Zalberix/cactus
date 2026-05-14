package main

import "testing"

func TestHTMLWorkerSourceIncludesVariantAndWorkerID(t *testing.T) {
	got := htmlWorkerSource("basic", 3)
	want := "worker/html-basic-3"
	if got != want {
		t.Fatalf("htmlWorkerSource() = %q, want %q", got, want)
	}
}

func TestHTMLWorkerNameUsesOverrideWhenProvided(t *testing.T) {
	got := htmlWorkerName("basic", "html-template-1")
	want := "html-template-1"
	if got != want {
		t.Fatalf("htmlWorkerName() = %q, want %q", got, want)
	}
}

func TestHTMLWorkerNameFallsBackToVariantName(t *testing.T) {
	got := htmlWorkerName("basic", "")
	want := "html-basic-worker"
	if got != want {
		t.Fatalf("htmlWorkerName() = %q, want %q", got, want)
	}
}
