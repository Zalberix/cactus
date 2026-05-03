package shell

import (
	"runtime"
)

const windowsOS = "windows"

func GetShell() string {
	if runtime.GOOS == windowsOS {
		return "cmd"
	} else {
		return "sh"
	}
}

func GetShellOption() string {
	if runtime.GOOS == windowsOS {
		return "/C"
	} else {
		return "-c"
	}
}
