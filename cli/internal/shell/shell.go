package shell

import (
	"runtime"
)

const windowsOS = "windows"

func GetShell() string {
	if runtime.GOOS == windowsOS {
		return "cmd"
	}
	return "sh"
}

func GetShellOption() string {
	if runtime.GOOS == windowsOS {
		return "/C"
	}
	return "-c"
}
