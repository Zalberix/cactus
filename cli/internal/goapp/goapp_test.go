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
