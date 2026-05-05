//go:build !windows

package kill

import (
	"context"
	"os/exec"
)

func killProcesses() error {
	return exec.CommandContext(context.Background(), "pkill", "-9", "-f", "cactus-").Run()
}
