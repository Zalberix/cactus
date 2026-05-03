package dependencies

import (
	"os"

	"github.com/zalberix/cactus/cli/internal/shell"
)

func installNodeDeps() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = shell.ExecCommand(
		shell.ExecCommandOpts{
			Command: "npm install --frozen-lockfile",
			Pwd:     wd,
			Stderr:  os.Stderr,
			Stdout:  os.Stdout,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
