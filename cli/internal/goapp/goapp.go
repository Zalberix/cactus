package goapp

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/zalberix/cactus/cli/internal/shell"
	"github.com/zalberix/cactus/cli/internal/watcher"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GoApp manages the lifecycle of a single Go microservice (build + run via dlv).
type GoApp struct {
	Name           string
	DebugPort      int
	Port           *int
	OnPortReady    func()
	Cmd            *exec.Cmd
	CorePath       string
	AppDir         string
	ConfigFileName string
	Watcher        *watcher.Watcher
	debugEnabled   bool
}

func NewApplication(name string, enableDebug bool, port *int, debugPort int, onPortReady func()) (
	*GoApp,
	error,
) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	app := GoApp{
		Name:         name,
		Cmd:          nil,
		CorePath:     wd,
		AppDir:       "apps",
		Watcher:      watcher.New(),
		debugEnabled: enableDebug,
		Port:         port,
		OnPortReady:  onPortReady,
		DebugPort:    debugPort,
	}

	cmd, err := app.CreateAppCommand()
	if err != nil {
		return nil, err
	}
	app.Cmd = cmd

	return &app, nil
}

func (g *GoApp) getBinAppPath() string {
	return filepath.Join(g.GetAppPath(), ".out", "cactus-"+g.Name)
}

func (g *GoApp) GetAppPath() string {
	return filepath.Join(g.CorePath, g.AppDir, g.Name)
}

func (g *GoApp) getConfigPath() string {
	return filepath.Join(g.CorePath, ".configs", g.AppDir, g.Name+".yaml")
}

// Build compiles the binary for this service.
func (g *GoApp) Build() error {
	pterm.Info.Printfln("[%s] Building...", g.Name)

	if err := os.MkdirAll(filepath.Dir(g.getBinAppPath()), 0o755); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}
	args := []string{"build"}
	if g.debugEnabled {
		args = append(args, `-gcflags=all=-N -l`)
	} else {
		args = append(args, `-ldflags=-s -w`)
	}
	args = append(args, "-o", g.getBinAppPath(), g.GetAppPath()+"/cmd/main.go")

	pterm.Info.Println("args for build:", args)

	cmd := exec.Command("go", args...)
	cmd.Dir = g.CorePath
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[%s] build: %w", g.Name, err)
	}

	pterm.Success.Printfln("[%s] Built", g.Name)
	return nil
}

func (g *GoApp) Stop() error {
	if g.Cmd == nil || g.Cmd.Process == nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		// /F - принудительно (Force)
		// /T - убить дерево процессов (Tree), то есть и dlv, и само приложение
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(g.Cmd.Process.Pid))
		err := killCmd.Run()
		if err != nil {
			return fmt.Errorf("taskkill failed: %w", err)
		}
		return nil
	}

	// Стандартное поведение для Linux/Mac
	return g.Cmd.Process.Kill()
}

func (g *GoApp) Start() error {
	if err := g.Stop(); err != nil {
		return fmt.Errorf("stoping error: %v", err)
	}

	if err := g.Build(); err != nil {
		return fmt.Errorf("building error: %v", err)
	}

	newCmd, err := g.CreateAppCommand()
	if err != nil {
		return fmt.Errorf("create app command: %v", err)
	}

	g.Cmd = newCmd

	go func() {
		g.waitPortReady()
	}()

	return g.Cmd.Start()
}

func (g *GoApp) CreateAppCommand() (*exec.Cmd, error) {
	// dlv exec .out/twir-emotes-cacher --headless=true --api-version=2 --check-go-version=false --only-same-user=false --listen=:2345 --log

	cmd, err := shell.CreateCommand(
		shell.ExecCommandOpts{
			Command: fmt.Sprintf(
				"dlv exec %s --headless=true --api-version=2 --check-go-version=false --only-same-user=false --listen=:%d --log --continue --accept-multiclient",
				g.getBinAppPath(),
				g.DebugPort,
			),
			Pwd:    g.CorePath,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
	)

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (g *GoApp) waitPortReady() {
	if g.Port == nil || g.OnPortReady == nil {
		return
	}

	for {
		_, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", *g.Port))
		if err != nil {
			continue
		}

		pterm.Info.Println("Port " + g.Name + " is ready, running hook")
		g.OnPortReady()
		break
	}
}
