package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/zalberix/cactus/cli/internal/cmds/dev"
	"github.com/zalberix/cactus/cli/internal/cmds/kill"
	"github.com/zalberix/cactus/cli/internal/cmds/linter"
	"github.com/zalberix/cactus/cli/internal/cmds/migrations"
	"github.com/zalberix/cactus/cli/internal/cmds/sqlcgen"
)

func main() {
	app := &cli.App{
		Name:  "cactus-cli",
		Usage: "Cactus development CLI",
		Commands: []*cli.Command{
			migrations.Command(),
			dev.Cmd,
			kill.Cmd,
			sqlcgen.Cmd,
			linter.LintCmd,
			linter.TestCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
