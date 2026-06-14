package frontend

import "testing"

func TestTargetSpecsReturnsEmptyWhenNoFrontendAppsConfigured(t *testing.T) {
	specs, err := TargetSpecs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 0 {
		t.Fatalf("expected no frontend target specs, got %#v", specs)
	}
}
