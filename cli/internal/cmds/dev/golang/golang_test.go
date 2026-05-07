package golang

import (
	"strings"
	"testing"

	"github.com/zalberix/cactus/cli/internal/goapp"
	"github.com/zalberix/cactus/cli/internal/services"
)

func TestExpandWorkerTemplatesUsesAppAndNamesByGroup(t *testing.T) {
	templates := []goapp.GoApp{
		{Name: "core", DebugPort: 2346, AppDir: "apps"},
		{Name: "smtp", IsWorker: true, DebugPort: 2348, AppDir: "apps/workers", DependsOn: []string{"core"}},
	}
	instances := []services.WorkerInstance{
		{Name: "smtp-basic", App: "smtp", Variant: "basic", UUID: "uuid-1", IDPath: ".worker_id/smtp-basic/uuid-1"},
		{Name: "smtp-auth", App: "smtp", Variant: "auth", UUID: "uuid-2", IDPath: ".worker_id/smtp-auth/uuid-2"},
		{Name: "smtp-auth", App: "smtp", Variant: "auth", UUID: "uuid-3", IDPath: ".worker_id/smtp-auth/uuid-3"},
	}

	expanded := expandWorkerTemplates(templates, instances)

	names := make([]string, 0, len(expanded))
	for _, app := range expanded {
		names = append(names, app.Name)
	}
	wantNames := []string{"core", "smtp-basic", "smtp-auth-1", "smtp-auth-2"}
	if strings.Join(names, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("expected names %v, got %v", wantNames, names)
	}

	for _, app := range expanded {
		if app.Name == "core" {
			continue
		}
		if app.BaseName != "smtp" {
			t.Fatalf("expected worker BaseName smtp, got %#v", app)
		}
		if app.WorkerGroupName == "" {
			t.Fatalf("expected WorkerGroupName to be set: %#v", app)
		}
		if app.WorkerVariant == "" {
			t.Fatalf("expected WorkerVariant to be set: %#v", app)
		}
		if len(app.ExtraArgs) != 6 {
			t.Fatalf("expected 6 ExtraArgs, got %#v", app.ExtraArgs)
		}
		if app.ExtraArgs[0] != "--worker-id-path" || app.ExtraArgs[2] != "--worker-variant" || app.ExtraArgs[4] != "--worker-name" {
			t.Fatalf("unexpected ExtraArgs: %#v", app.ExtraArgs)
		}
	}
}

func TestExpandWorkerTemplatesPassesUniqueRuntimeWorkerNames(t *testing.T) {
	templates := []goapp.GoApp{
		{Name: "smtp", IsWorker: true, DebugPort: 2348, AppDir: "apps/workers"},
	}
	instances := []services.WorkerInstance{
		{Name: "smtp-basic", App: "smtp", Variant: "basic", UUID: "uuid-1", IDPath: ".worker_id/smtp-basic/uuid-1"},
		{Name: "smtp-basic", App: "smtp", Variant: "basic", UUID: "uuid-2", IDPath: ".worker_id/smtp-basic/uuid-2"},
		{Name: "smtp-basic", App: "smtp", Variant: "basic", UUID: "uuid-3", IDPath: ".worker_id/smtp-basic/uuid-3"},
	}

	expanded := expandWorkerTemplates(templates, instances)

	got := []string{
		extraArgValue(t, expanded[0].ExtraArgs, "--worker-name"),
		extraArgValue(t, expanded[1].ExtraArgs, "--worker-name"),
		extraArgValue(t, expanded[2].ExtraArgs, "--worker-name"),
	}
	want := []string{"smtp-basic-1", "smtp-basic-2", "smtp-basic-3"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected runtime worker names %v, got %v", want, got)
	}
}

func TestExpandWorkerTemplatesAssignsUniqueDebugPortsAcrossSameApp(t *testing.T) {
	templates := []goapp.GoApp{
		{Name: "smtp", IsWorker: true, DebugPort: 2348, AppDir: "apps/workers"},
	}
	instances := []services.WorkerInstance{
		{Name: "smtp-basic", App: "smtp", Variant: "basic", UUID: "uuid-1", IDPath: ".worker_id/smtp-basic/uuid-1"},
		{Name: "smtp-auth", App: "smtp", Variant: "auth", UUID: "uuid-2", IDPath: ".worker_id/smtp-auth/uuid-2"},
		{Name: "smtp-auth", App: "smtp", Variant: "auth", UUID: "uuid-3", IDPath: ".worker_id/smtp-auth/uuid-3"},
	}

	expanded := expandWorkerTemplates(templates, instances)

	ports := []int{expanded[0].DebugPort, expanded[1].DebugPort, expanded[2].DebugPort}
	want := []int{2348, 2349, 2350}
	for i := range want {
		if ports[i] != want[i] {
			t.Fatalf("expected debug ports %v, got %v", want, ports)
		}
	}
}

func extraArgValue(t *testing.T, args []string, name string) string {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1]
		}
	}
	t.Fatalf("missing %s in %#v", name, args)
	return ""
}
