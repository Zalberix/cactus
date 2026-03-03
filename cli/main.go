package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"cactus/cli/internal/cmds/dev"
	"cactus/cli/internal/cmds/kill"
	"cactus/cli/internal/cmds/migrations"
)

func main() {
	app := &cli.App{
		Name:  "cactus-cli",
		Usage: "Cactus development CLI",
		Commands: []*cli.Command{
			migrations.Command(),
			dev.Cmd,
			kill.Cmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
