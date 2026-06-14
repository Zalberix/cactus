package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
	"github.com/zalberix/cactus/cli/internal/shell"

	"github.com/pterm/pterm"
)

// StartProxy launches Caddy via `go tool caddy` using the project Caddyfile.
// It blocks until port 80 is reachable or times out.
// Pass debug=true to enable Caddy's --watch flag.
func StartProxy(_ bool) (<-chan struct{}, error) {
	pterm.Info.Println("Starting Caddy proxy...")
	startChannel := make(chan struct{})

	wd, err := os.Getwd()
	if err != nil {
		return startChannel, err
	}

	go func() {
		for !checkIsProxyStarted(80) {
			pterm.Info.Println("Waiting for proxy to start")
			time.Sleep(500 * time.Millisecond)
		}

		pterm.Success.Println("Proxy started")
		startChannel <- struct{}{}
		close(startChannel)
	}()

	commandOpts := shell.ExecCommandOpts{
		Command: "go tool caddy run --watch --config Caddyfile",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Pwd:     wd,
	}

	go func() {
		err = shell.ExecCommand(commandOpts)
		if err != nil {
			panic(err)
		}
	}()

	return startChannel, err
}

func TargetSpec() devruntime.TargetSpec {
	return devruntime.TargetSpec{
		ID:              "caddy-proxy",
		Name:            "caddy-proxy",
		Kind:            devruntime.TargetProcess,
		GracefulTimeout: 5 * time.Second,
		Command: func(ctx context.Context, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
			wd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			return shell.CreateCommand(shell.ExecCommandOpts{
				Context: ctx,
				Command: "go tool caddy run --watch --config Caddyfile",
				Stdout:  stdout,
				Stderr:  stderr,
				Pwd:     wd,
			})
		},
		Ready: devruntime.PortReadyProbe(80, 500*time.Millisecond, 60*time.Second),
	}
}

func checkIsProxyStarted(port int) bool {
	_, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", fmt.Sprintf("127.0.0.1:%d", port))
	return err == nil
}
