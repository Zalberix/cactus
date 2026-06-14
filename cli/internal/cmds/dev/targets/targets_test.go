package targets

import (
	"testing"

	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

func TestBuildTargetsIncludesStartupProxyAndGoApps(t *testing.T) {
	specs, err := BuildTargets(devruntime.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) < 5 {
		t.Fatalf("expected startup targets and go apps, got %#v", specs)
	}

	wantPrefix := []string{"clean-ports", "caddy-proxy", "deps", "migrations"}
	for i, want := range wantPrefix {
		if specs[i].ID != want {
			t.Fatalf("spec[%d].ID = %q, want %q; specs=%#v", i, specs[i].ID, want, specs)
		}
	}
	if specs[2].Skip != true || specs[3].Skip != true {
		t.Fatalf("deps and migrations should be skipped by default: %#v", specs[:4])
	}

	var foundManager bool
	for _, spec := range specs {
		if spec.ID == "manager" && spec.Kind == devruntime.TargetProcess {
			foundManager = true
		}
	}
	if !foundManager {
		t.Fatalf("manager target not found in %#v", specs)
	}
}
