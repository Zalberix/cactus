//go:build !windows

package kill

import "os/exec"

func killProcesses() error {
	return exec.Command("pkill", "-9", "-f", "twir-").Run()
}
