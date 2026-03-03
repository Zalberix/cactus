package frontend

import (
	"os"
	"path/filepath"
)

// Apps lists all frontend applications managed by the CLI.
var Apps = []app{
	{name: "web", path: "web"},
}

type FrontendApps struct {
	apps []*app
}

func New() (*FrontendApps, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fa := &FrontendApps{}
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

func (fa *FrontendApps) Start() error {
	for _, app := range fa.apps {
		if err := app.start(); err != nil {
			return err
		}
	}

	return nil
}

func (fa *FrontendApps) Stop() error {
	for _, app := range fa.apps {
		if err := app.stop(); err != nil {
			return err
		}
	}

	return nil
}
