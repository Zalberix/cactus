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
		{Name: "smtp-rich", App: "smtp", Variant: "rich", UUID: "uuid-2", IDPath: ".worker_id/smtp-rich/uuid-2"},
		{Name: "smtp-rich", App: "smtp", Variant: "rich", UUID: "uuid-3", IDPath: ".worker_id/smtp-rich/uuid-3"},
	}

	expanded := expandWorkerTemplates(templates, instances)

	names := make([]string, 0, len(expanded))
	for _, app := range expanded {
		names = append(names, app.Name)
	}
	wantNames := []string{"core", "smtp-basic", "smtp-rich-1", "smtp-rich-2"}
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
		if len(app.ExtraArgs) != 4 {
			t.Fatalf("expected 4 ExtraArgs, got %#v", app.ExtraArgs)
		}
		if app.ExtraArgs[0] != "--worker-id-path" || app.ExtraArgs[2] != "--worker-variant" {
			t.Fatalf("unexpected ExtraArgs: %#v", app.ExtraArgs)
		}
	}
}

func TestExpandWorkerTemplatesAssignsUniqueDebugPortsAcrossSameApp(t *testing.T) {
	templates := []goapp.GoApp{
		{Name: "smtp", IsWorker: true, DebugPort: 2348, AppDir: "apps/workers"},
	}
	instances := []services.WorkerInstance{
		{Name: "smtp-basic", App: "smtp", Variant: "basic", UUID: "uuid-1", IDPath: ".worker_id/smtp-basic/uuid-1"},
		{Name: "smtp-rich", App: "smtp", Variant: "rich", UUID: "uuid-2", IDPath: ".worker_id/smtp-rich/uuid-2"},
		{Name: "smtp-rich", App: "smtp", Variant: "rich", UUID: "uuid-3", IDPath: ".worker_id/smtp-rich/uuid-3"},
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
