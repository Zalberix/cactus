package goapp

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

func TestBuildArgsTargetCommandPackageDirectory(t *testing.T) {
	app := GoApp{
		Name:     "smtp",
		CorePath: filepath.Clean(`S:\repo\cactus`),
		AppDir:   filepath.Join("apps", "workers"),
	}

	args := app.buildArgs()
	target := args[len(args)-1]
	want := filepath.Join(app.GetAppPath(), "cmd")
	if target != want {
		t.Fatalf("expected build target %q, got %q", want, target)
	}
}

func TestAppsUsesTemplateWorkerName(t *testing.T) {
	for _, app := range Apps {
		if app.Name == "template" && app.IsWorker {
			return
		}
		if (app.Name == "html" || app.Name == "template-html") && app.IsWorker {
			t.Fatal("HTML template worker app should be named template")
		}
	}
	t.Fatal("template worker app is not registered")
}

func TestBuildArgsUsesCustomCommandDir(t *testing.T) {
	app := GoApp{
		Name:       "core-worker",
		CorePath:   filepath.Clean(`S:\repo\cactus`),
		SourceDir:  filepath.Join("apps", "core"),
		CommandDir: filepath.Join("apps", "core", "cmd"),
	}

	args := app.buildArgs()
	target := args[len(args)-1]
	want := filepath.Join(app.CorePath, "apps", "core", "cmd")
	if target != want {
		t.Fatalf("expected build target %q, got %q", want, target)
	}
}

func TestCreateAppCommandWithInjectsContextWritersAndArgs(t *testing.T) {
	app := GoApp{
		Name:      "manager",
		CorePath:  t.TempDir(),
		AppDir:    "apps",
		ExtraArgs: []string{"--config", "local"},
	}
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd, err := app.CreateAppCommandWith(context.Background(), stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}

	if cmd.Stdout != stdout {
		t.Fatal("stdout writer was not injected")
	}
	if cmd.Stderr != stderr {
		t.Fatal("stderr writer was not injected")
	}
	if cmd.Dir != app.CorePath {
		t.Fatalf("cmd.Dir = %q, want %q", cmd.Dir, app.CorePath)
	}
	if got := cmd.Args[len(cmd.Args)-2:]; got[0] != "--config" || got[1] != "local" {
		t.Fatalf("extra args were not preserved: %#v", cmd.Args)
	}
}

func TestAppsRegistersManagerFromManagerSource(t *testing.T) {
	for _, app := range Apps {
		if app.Name != "manager" {
			continue
		}
		if app.AppDir != "apps" {
			t.Fatalf("manager must use apps app dir, got %q", app.AppDir)
		}
		if app.SourceDir != "" {
			t.Fatalf("manager must use default apps/manager source dir, got %q", app.SourceDir)
		}
		if app.CommandDir != "" {
			t.Fatalf("manager must use default apps/manager/cmd command dir, got %q", app.CommandDir)
		}
		if app.Port == nil {
			t.Fatal("manager must keep the HTTP port")
		}
		return
	}
	t.Fatal("manager app is not registered")
}

func TestAppsRegistersCoreWorkerOutsideBusinessWorkers(t *testing.T) {
	for _, app := range Apps {
		if app.Name != "core-worker" {
			continue
		}
		if app.IsWorker {
			t.Fatal("core-worker must not use business worker expansion")
		}
		if !app.IsCoreWorker {
			t.Fatal("core-worker must be marked as core orchestration worker")
		}
		if app.SourceDir != filepath.Join("apps", "core") {
			t.Fatalf("core-worker must use apps/core source dir, got %q", app.SourceDir)
		}
		if app.CommandDir != filepath.Join("apps", "core", "cmd") {
			t.Fatalf("core-worker must build from apps/core/cmd, got %q", app.CommandDir)
		}
		if len(app.DependsOn) != 1 || app.DependsOn[0] != "manager" {
			t.Fatalf("core-worker must depend on manager, got %#v", app.DependsOn)
		}
		return
	}
	t.Fatal("core-worker app is not registered")
}
