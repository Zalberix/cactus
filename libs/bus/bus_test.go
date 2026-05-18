package bus

import "testing"

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
