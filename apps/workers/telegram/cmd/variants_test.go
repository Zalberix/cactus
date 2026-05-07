package main

import (
	"testing"

	"github.com/zalberix/cactus/apps/workers/telegram/config"
	"github.com/zalberix/cactus/libs/worker"
)

func TestTelegramVariantsExposeBasicManifest(t *testing.T) {
	variants := telegramVariants(config.Config{})

	basic, err := worker.SelectVariant(variants, "basic")
	if err != nil {
		t.Fatalf("select basic: %v", err)
	}

	if basic.Manifest.Kind != "telegram-basic" {
		t.Fatalf("expected basic kind telegram-basic, got %q", basic.Manifest.Kind)
	}
	if basic.Manifest.Type != "social" {
		t.Fatalf("expected social worker type, got %q", basic.Manifest.Type)
	}
	if basic.Handler == nil {
		t.Fatal("expected basic variant handler")
	}
}

func TestTelegramVariantsRejectUnknownVariant(t *testing.T) {
	_, err := worker.SelectVariant(telegramVariants(config.Config{}), "missing")
	if err == nil {
		t.Fatal("expected unknown variant error")
	}
}

func TestTelegramWorkerSourceIncludesVariantAndWorkerID(t *testing.T) {
	got := telegramWorkerSource("basic", 3)
	want := "worker/telegram-basic-3"
	if got != want {
		t.Fatalf("telegramWorkerSource() = %q, want %q", got, want)
	}
}

func TestTelegramWorkerNameUsesOverrideWhenProvided(t *testing.T) {
	got := telegramWorkerName("basic", "telegram-basic-2")
	want := "telegram-basic-2"
	if got != want {
		t.Fatalf("telegramWorkerName() = %q, want %q", got, want)
	}
}

func TestTelegramWorkerNameFallsBackToVariantName(t *testing.T) {
	got := telegramWorkerName("basic", "")
	want := "telegram-basic-worker"
	if got != want {
		t.Fatalf("telegramWorkerName() = %q, want %q", got, want)
	}
}
