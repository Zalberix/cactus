package proxy

import (
	"bytes"
	"context"
	"strings"
	"testing"

	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

func TestTargetSpecCreatesManagedCaddyProcess(t *testing.T) {
	spec := TargetSpec()
	if spec.ID != "caddy-proxy" || spec.Name != "caddy-proxy" {
		t.Fatalf("unexpected proxy identity: %#v", spec)
	}
	if spec.Kind != devruntime.TargetProcess {
		t.Fatalf("kind = %q, want %q", spec.Kind, devruntime.TargetProcess)
	}
	if spec.Command == nil || spec.Ready == nil {
		t.Fatalf("proxy target must include command and readiness: %#v", spec)
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd, err := spec.Command(context.Background(), stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Stdout != stdout || cmd.Stderr != stderr {
		t.Fatal("command did not use injected writers")
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "go tool caddy run --watch --config Caddyfile") {
		t.Fatalf("unexpected command args: %#v", cmd.Args)
	}
}
