package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempConfig(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "cactus-services.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadServicesParsesWorkerList(t *testing.T) {
	path := writeTempConfig(t, `
workers:
  - name: smtp-basic
    app: smtp
    variant: basic
    count: 1
  - name: smtp-auth
    app: smtp
    variant: auth
    count: 2
`)

	cfg, err := LoadServices(path)
	if err != nil {
		t.Fatalf("LoadServices returned error: %v", err)
	}

	if len(cfg.Workers) != 2 {
		t.Fatalf("expected 2 workers, got %d", len(cfg.Workers))
	}

	first := cfg.Workers[0]
	if first.Name != "smtp-basic" || first.App != "smtp" || first.Variant != "basic" || first.Count != 1 {
		t.Fatalf("unexpected first worker: %#v", first)
	}

	second := cfg.Workers[1]
	if second.Name != "smtp-auth" || second.App != "smtp" || second.Variant != "auth" || second.Count != 2 {
		t.Fatalf("unexpected second worker: %#v", second)
	}
}

func TestLoadServicesRejectsLegacyMapFormat(t *testing.T) {
	path := writeTempConfig(t, `
workers:
  smtp:
    count: 1
`)

	_, err := LoadServices(path)
	if err == nil {
		t.Fatal("expected legacy map format to fail")
	}
	if !strings.Contains(err.Error(), "parse services config") {
		t.Fatalf("expected parse error, got: %v", err)
	}
}

func TestLoadServicesValidatesWorkerFields(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "empty name",
			body: `
workers:
  - name: ""
    app: smtp
    variant: basic
    count: 1
`,
			want: "workers[0].name is required",
		},
		{
			name: "empty app",
			body: `
workers:
  - name: smtp-basic
    app: ""
    variant: basic
    count: 1
`,
			want: "workers[0].app is required",
		},
		{
			name: "empty variant",
			body: `
workers:
  - name: smtp-basic
    app: smtp
    variant: ""
    count: 1
`,
			want: "workers[0].variant is required",
		},
		{
			name: "negative count",
			body: `
workers:
  - name: smtp-basic
    app: smtp
    variant: basic
    count: -1
`,
			want: "workers[0].count must be >= 0",
		},
		{
			name: "duplicate name",
			body: `
workers:
  - name: smtp-basic
    app: smtp
    variant: basic
    count: 1
  - name: smtp-basic
    app: smtp
    variant: auth
    count: 1
`,
			want: `duplicate worker name "smtp-basic"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempConfig(t, tc.body)

			_, err := LoadServices(path)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestReconcileUsesWorkerNameForLockAndRuntimeID(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "cactus-services.yaml")
	lockPath := filepath.Join(dir, "cactus-services-lock.yaml")
	workerIDDir := filepath.Join(dir, ".worker_id")

	if err := os.WriteFile(cfgPath, []byte(`
workers:
  - name: smtp-basic
    app: smtp
    variant: basic
    count: 1
  - name: smtp-auth
    app: smtp
    variant: auth
    count: 2
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	instances, err := Reconcile(cfgPath, lockPath, workerIDDir)
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if len(instances) != 3 {
		t.Fatalf("expected 3 instances, got %d", len(instances))
	}

	counts := map[string]int{}
	for _, inst := range instances {
		counts[inst.Name]++
		if inst.App != "smtp" {
			t.Fatalf("expected app smtp, got %#v", inst)
		}
		if inst.Name == "smtp-basic" && inst.Variant != "basic" {
			t.Fatalf("expected smtp-basic variant basic, got %#v", inst)
		}
		if inst.Name == "smtp-auth" && inst.Variant != "auth" {
			t.Fatalf("expected smtp-auth variant auth, got %#v", inst)
		}
		if !strings.Contains(filepath.ToSlash(inst.IDPath), "/.worker_id/"+inst.Name+"/") {
			t.Fatalf("expected IDPath to include group name, got %q", inst.IDPath)
		}
	}
	if counts["smtp-basic"] != 1 || counts["smtp-auth"] != 2 {
		t.Fatalf("unexpected counts: %#v", counts)
	}

	lockData, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	lockText := string(lockData)
	if !strings.Contains(lockText, "smtp-basic:") || !strings.Contains(lockText, "smtp-auth:") {
		t.Fatalf("expected lock keyed by worker names, got:\n%s", lockText)
	}
}

func TestReconcileRemovesDeletedWorkerGroups(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "cactus-services.yaml")
	lockPath := filepath.Join(dir, "cactus-services-lock.yaml")
	workerIDDir := filepath.Join(dir, ".worker_id")

	if err := os.WriteFile(cfgPath, []byte(`
workers:
  - name: smtp-basic
    app: smtp
    variant: basic
    count: 1
`), 0o600); err != nil {
		t.Fatalf("write first config: %v", err)
	}
	if _, err := Reconcile(cfgPath, lockPath, workerIDDir); err != nil {
		t.Fatalf("first reconcile: %v", err)
	}

	if err := os.WriteFile(cfgPath, []byte(`
workers:
  - name: smtp-auth
    app: smtp
    variant: auth
    count: 1
`), 0o600); err != nil {
		t.Fatalf("write second config: %v", err)
	}

	instances, err := Reconcile(cfgPath, lockPath, workerIDDir)
	if err != nil {
		t.Fatalf("second reconcile: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].Name != "smtp-auth" {
		t.Fatalf("expected remaining group smtp-auth, got %#v", instances[0])
	}

	lockData, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	lockText := string(lockData)
	if strings.Contains(lockText, "smtp-basic:") {
		t.Fatalf("expected deleted group to be removed from lock, got:\n%s", lockText)
	}
}
