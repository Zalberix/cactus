//go:build windows

package kill

import "os/exec"

func killProcesses() error {
	return exec.Command(
		"powershell", "-c",
		`Get-Process | Where-Object {$_.ProcessName -like 'twir-*'} | Stop-Process -Force`,
	).Run()
}
