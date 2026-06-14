package goapp

import "path/filepath"

var managerPort = 3009

// Apps is the canonical list of Go services in the monorepo.
var Apps = []GoApp{
	{
		Name:      "manager",
		DebugPort: 2346,
		AppDir:    "apps",
		Port:      &managerPort,
	},
	{
		Name:         "core-worker",
		DebugPort:    2350,
		SourceDir:    filepath.Join("apps", "core"),
		CommandDir:   filepath.Join("apps", "core", "cmd"),
		IsCoreWorker: true,
		DependsOn:    []string{"manager"},
	},
	// {Name: "telegram", IsWorker: true, DebugPort: 2347, AppDir: "apps/workers", DependsOn: []string{"manager"}},
	{Name: "smtp", IsWorker: true, DebugPort: 2348, AppDir: "apps/workers", DependsOn: []string{"manager"}},
	{Name: "template", IsWorker: true, DebugPort: 2349, AppDir: "apps/workers", DependsOn: []string{"manager"}},
}
