package sqlcgen

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/zalberix/cactus/cli/internal/shell"
)

var Cmd = &cli.Command{
	Name:  "sqlc-generate",
	Usage: "Generate Go code from SQL queries via sqlc",
	Action: func(_ *cli.Context) error {
		pterm.Info.Println("Running sqlc generate...")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = shell.ExecCommand(shell.ExecCommandOpts{
			Command: "sqlc generate",
			Pwd:     wd,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		})
		if err != nil {
			pterm.Error.Printfln("sqlc generate failed: %v", err)
			return err
		}

		pterm.Success.Println("sqlc generate complete")
		return nil
	},
}
