package bus

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestNewStreamConfigAppliesOptions(t *testing.T) {
	cfg := newStreamConfig("CONFIGS", []string{"config.>"}, WithAllowDirect())

	if cfg.Name != "CONFIGS" {
		t.Fatalf("Name = %q, want CONFIGS", cfg.Name)
	}
	if len(cfg.Subjects) != 1 || cfg.Subjects[0] != "config.>" {
		t.Fatalf("Subjects = %#v, want config.>", cfg.Subjects)
	}
	if !cfg.AllowDirect {
		t.Fatal("expected AllowDirect to be enabled")
	}
}

func TestNATSConnectOptionsRetryForeverEverySecond(t *testing.T) {
	opts := nats.GetDefaultOptions()
	for _, opt := range natsConnectOptions(Options{URL: "tls://localhost:4222"}) {
		if err := opt(&opts); err != nil {
			t.Fatalf("apply option: %v", err)
		}
	}

	if opts.MaxReconnect != -1 {
		t.Fatalf("MaxReconnect = %d, want -1", opts.MaxReconnect)
	}
	if opts.ReconnectWait != time.Second {
		t.Fatalf("ReconnectWait = %s, want 1s", opts.ReconnectWait)
	}
}
