package frontend

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

// Apps lists all frontend applications managed by the CLI.
// Frontend (Nuxt.js) will be added in Phase 5.
var Apps = []app{}

type AppsGroup struct {
	apps []*app
}

func New() (*AppsGroup, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fa := &AppsGroup{}
	for _, app := range Apps {
		path := filepath.Join(wd, app.path)
		application, err := newApplication(app.name, path)
		if err != nil {
			return nil, err
		}

		fa.apps = append(fa.apps, application)
	}

	return fa, nil
}

func (fa *AppsGroup) Start() error {
	for _, app := range fa.apps {
		if err := app.start(); err != nil {
			return err
		}
	}

	return nil
}

func (fa *AppsGroup) Stop() error {
	for _, app := range fa.apps {
		if err := app.stop(); err != nil {
			return err
		}
	}

	return nil
}

func TargetSpecs() ([]devruntime.TargetSpec, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	specs := make([]devruntime.TargetSpec, 0, len(Apps))
	for _, configured := range Apps {
		configured := configured
		path := filepath.Join(wd, configured.path)

		specs = append(specs, devruntime.TargetSpec{
			ID:                  configured.name,
			Name:                configured.name,
			Kind:                devruntime.TargetProcess,
			GracefulTimeout:     5 * time.Second,
			RestartOnFileChange: true,
			WatchPaths:          []string{path},
			Command: func(ctx context.Context, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
				application := app{name: configured.name, path: path}
				return application.createAppCommandWith(ctx, stdout, stderr)
			},
			Ready: devruntime.ImmediateReady(),
		})
	}

	return specs, nil
}
