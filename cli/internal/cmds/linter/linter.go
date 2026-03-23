package linter

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/zalberix/cactus/cli/internal/shell"
)

var LintCmd = &cli.Command{
	Name:  "lint",
	Usage: "Run golangci-lint on all modules",
	Action: func(c *cli.Context) error {
		pterm.Info.Println("Running golangci-lint...")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = shell.ExecCommand(shell.ExecCommandOpts{
			Command: "golangci-lint run ./...",
			Pwd:     wd,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		})
		if err != nil {
			pterm.Error.Printfln("Lint failed: %v", err)
			return err
		}

		pterm.Success.Println("Lint passed")
		return nil
	},
}

var TestCmd = &cli.Command{
	Name:  "test",
	Usage: "Run go test on all modules",
	Action: func(c *cli.Context) error {
		pterm.Info.Println("Running tests...")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = shell.ExecCommand(shell.ExecCommandOpts{
			Command: "go test ./...",
			Pwd:     wd,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		})
		if err != nil {
			pterm.Error.Printfln("Tests failed: %v", err)
			return err
		}

		pterm.Success.Println("All tests passed")
		return nil
	},
}
