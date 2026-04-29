package proxy

import (
	"fmt"
	"github.com/zalberix/cactus/cli/internal/shell"
	"net"
	"os"
	"time"

	"github.com/pterm/pterm"
)

// StartProxy launches Caddy via `go tool caddy` using the project Caddyfile.
// It blocks until port 80 is reachable or times out.
// Pass debug=true to enable Caddy's --watch flag.
func StartProxy(debug bool) (<-chan struct{}, error) {
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

func checkIsProxyStarted(port int) bool {
	_, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	return err == nil
}
