// Package gum wraps the Gum CLI for non-fullscreen command flows.
// Do not execute Gum while a Bubble Tea program is active because both tools
// need control of terminal input and output.
package gum

import (
	"context"
	"os/exec"
	"strings"
)

func Available(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "go", "tool", "gum", "--version")
	return cmd.Run() == nil
}

func FormatCommand(args ...string) string {
	command := "go tool gum"
	for _, arg := range args {
		command += " " + shellQuote(arg)
	}
	return command
}

func shellQuote(value string) string {
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	return `"` + escaped + `"`
}
