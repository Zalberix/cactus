package helpers

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/zalberix/cactus/cli/internal/goapp"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var CleanPortsCmd = &cli.Command{
	Name:  "clean-ports",
	Usage: "Kill all processes listening on development ports",
	Action: func(_ *cli.Context) error {
		// Главное не задеть порты docker, тк он зависает
		ports := []int{80, 3010, 3009}
		for _, app := range goapp.Apps {
			ports = append(ports, app.DebugPort)
		}

		pterm.Info.Printfln("Cleaning %d ports...", len(ports))
		for _, port := range ports {
			if err := killPort(port); err != nil {
				pterm.Warning.Printfln("Port %d: %v", port, err)
			} else {
				pterm.Success.Printfln("Port %d cleared", port)
			}
		}
		return nil
	},
}

func killPort(port int) error {
	if runtime.GOOS == "windows" {
		return killPortWindows(port)
	}
	return killPortUnix(port)
}

func killPortUnix(port int) error {
	out, err := exec.CommandContext(context.Background(), "lsof", "-ti", fmt.Sprintf(":%d", port)).Output() // #nosec G204 -- port is selected by the CLI from known dev ports.
	if err != nil {
		// no process on this port — not an error
		return nil //nolint:nilerr
	}
	pids := strings.Fields(strings.TrimSpace(string(out)))
	for _, pid := range pids {
		_ = exec.CommandContext(context.Background(), "kill", "-9", pid).Run()
	}
	return nil
}

func killPortWindows(port int) error {
	script := fmt.Sprintf(
		`$c = Get-NetTCPConnection -LocalPort %d -ErrorAction SilentlyContinue; `+
			`if ($c) { Stop-Process -Id $c.OwningProcess -Force }`,
		port,
	)
	return exec.CommandContext(context.Background(), "powershell", "-c", script).Run()
}
