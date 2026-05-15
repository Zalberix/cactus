package goapp

import (
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
