package build

import (
	"cactus/cli/internal/shell"

	"github.com/pterm/pterm"
)

// LibsCmd builds all JS libraries with npm.
func LibsCmd(rootDir string) error {
	pterm.Info.Println("Building JS libs...")
	return shell.ExecCommand(shell.ExecCommandOpts{
		Command: "npm run build --workspace=@cactus/web",
		Pwd:     rootDir,
	})
}
