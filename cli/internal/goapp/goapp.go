package goapp

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pterm/pterm"
	"github.com/zalberix/cactus/cli/internal/shell"
	"github.com/zalberix/cactus/cli/internal/watcher"
)

// GoApp manages the lifecycle of a single Go microservice (build + run via dlv).
type GoApp struct {
	Name         string
	DebugPort    int
	Port         *int
	OnPortReady  func()
	DependsOn    []string
	Ready        chan struct{}
	Cmd          *exec.Cmd
	CorePath     string
	AppDir       string
	Watcher      *watcher.Watcher
	debugEnabled bool

	// IsWorker — true для worker-сервисов, которые масштабируются через cactus-services.yaml.
	IsWorker bool
	// WorkerUUID — UUID инстанса из lock-файла. Заполняется при expansion.
	WorkerUUID string
	// BaseName — оригинальное имя приложения (например "smtp") до expansion в "smtp-1".
	BaseName string
	// ExtraArgs — дополнительные аргументы для бинарника (передаются после -- в dlv).
	ExtraArgs []string
}

func NewApplication(name string, enableDebug bool, appDir string, port *int, debugPort int, onPortReady func(), dependsOn []string) (
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
		AppDir:       appDir,
		Watcher:      watcher.New(),
		debugEnabled: enableDebug,
		Port:         port,
		OnPortReady:  onPortReady,
		DebugPort:    debugPort,
		DependsOn:    dependsOn,
	}

	cmd, err := app.CreateAppCommand()
	if err != nil {
		return nil, err
	}
	app.Cmd = cmd

	return &app, nil
}

func (g *GoApp) getBinAppPath() string {
	name := g.Name
	if g.BaseName != "" {
		name = g.BaseName
	}
	binName := "cactus-" + name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return filepath.Join(g.GetAppPath(), ".out", binName)
}

func (g *GoApp) GetAppPath() string {
	name := g.Name
	if g.BaseName != "" {
		name = g.BaseName
	}
	return filepath.Join(g.CorePath, g.AppDir, name)
}

func (g *GoApp) getConfigPath() string {
	name := g.Name
	if g.BaseName != "" {
		name = g.BaseName
	}
	return filepath.Join(g.CorePath, "configs", g.AppDir, name+".yaml")
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
		if err := killCmd.Run(); err != nil {
			// Exit code 128 — процесс уже завершён, не ошибка
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 128 {
				return nil
			}
			return fmt.Errorf("taskkill failed: %w", err)
		}
		return nil
	}

	// Стандартное поведение для Linux/Mac
	return g.Cmd.Process.Kill()
}

func (g *GoApp) Start() error {
	if err := g.Stop(); err != nil {
		return fmt.Errorf("stoping error: %w", err)
	}

	if err := g.Build(); err != nil {
		return fmt.Errorf("building error: %w", err)
	}

	newCmd, err := g.CreateAppCommand()
	if err != nil {
		return fmt.Errorf("create app command: %w", err)
	}

	g.Cmd = newCmd

	// Создаём новый Ready канал при каждом (ре)старте
	g.Ready = make(chan struct{})

	if err := g.Cmd.Start(); err != nil {
		return err
	}

	go g.signalReady()

	return nil
}

// signalReady закрывает канал Ready когда приложение готово:
// если Port задан — ждёт ответа порта, иначе — сразу.
func (g *GoApp) signalReady() {
	defer func() {
		select {
		case <-g.Ready:
			// уже закрыт
		default:
			close(g.Ready)
		}
	}()

	if g.Port == nil {
		return
	}

	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", *g.Port))
		if err == nil {
			conn.Close()
			pterm.Success.Printfln("[%s] Порт %d готов", g.Name, *g.Port)
			if g.OnPortReady != nil {
				g.OnPortReady()
			}
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func (g *GoApp) CreateAppCommand() (*exec.Cmd, error) {
	if !g.debugEnabled {
		cmd := exec.Command(g.getBinAppPath(), g.ExtraArgs...)
		cmd.Dir = g.CorePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd, nil
	}

	command := fmt.Sprintf(
		"dlv exec %s --headless=true --api-version=2 --check-go-version=false --only-same-user=false --listen=:%d --log --continue --accept-multiclient",
		g.getBinAppPath(),
		g.DebugPort,
	)

	if len(g.ExtraArgs) > 0 {
		command += " --"
		for _, arg := range g.ExtraArgs {
			command += " " + arg
		}
	}

	cmd, err := shell.CreateCommand(
		shell.ExecCommandOpts{
			Command: command,
			Pwd:     g.CorePath,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
