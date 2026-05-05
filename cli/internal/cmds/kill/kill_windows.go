//go:build windows

package kill

import (
	"context"
	"os/exec"
)

func killProcesses() error {
	return exec.CommandContext(
		context.Background(),
		"powershell", "-c",
		`Get-Process | Where-Object {$_.ProcessName -like 'cactus-*'} | Stop-Process -Force`,
	).Run()
}
