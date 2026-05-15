package main

import "testing"

func TestHTMLWorkerSourceIncludesVariantAndWorkerID(t *testing.T) {
	got := templateWorkerSource("html", 3)
	want := "worker/template-html-3"
	if got != want {
		t.Fatalf("templateWorkerSource() = %q, want %q", got, want)
	}
}

func TestHTMLWorkerNameUsesOverrideWhenProvided(t *testing.T) {
	got := templateWorkerName("html", "template-html-1")
	want := "template-html-1"
	if got != want {
		t.Fatalf("templateWorkerName() = %q, want %q", got, want)
	}
}

func TestHTMLWorkerNameFallsBackToVariantName(t *testing.T) {
	got := templateWorkerName("html", "")
	want := "template-html-worker"
	if got != want {
		t.Fatalf("templateWorkerName() = %q, want %q", got, want)
	}
}
