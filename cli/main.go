package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"cactus/cli/internal/cmds/migrations"
)

func main() {
	app := &cli.App{
		Name:  "cactus-cli",
		Usage: "Cactus development CLI",
		Commands: []*cli.Command{
			migrations.Command(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
