package goapp

var corePort = 3009

// Apps is the canonical list of Go services in the monorepo.
var Apps = []GoApp{
	{Name: "core", DebugPort: 2346, AppDir: "apps", Port: &corePort},
	//{Name: "telegram", DebugPort: 2347, AppDir: "apps/workers", DependsOn: []string{"core"}},
	{Name: "smtp", DebugPort: 2348, AppDir: "apps/workers", DependsOn: []string{"core"}},
}
