package goapp

// App describes a Go microservice managed by the CLI.
type App struct {
	Name        string
	Port        int    // HTTP port (0 if none)
	DebugPort   int    // Delve debug port
	OnPortReady func() // called once the HTTP port is reachable (optional)
}

// Apps is the canonical list of Go services in the monorepo.
var Apps = []GoApp{
	{Name: "core", DebugPort: 2346, ConfigFileName: "configs/apps/core.yaml", AppDir: "core"},
	//{Name: "telegram", DebugPort: 2347, ConfigFileName: "configs/apps/workers/telegram.yaml", AppDir: "workers/telegram"},
	//{Name: "smtp", DebugPort: 2348, ConfigFileName: "configs/apps/workers/smtp.yaml", AppDir: "workers/smtp"},
}
