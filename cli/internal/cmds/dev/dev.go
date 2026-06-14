package dev

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/zalberix/cactus/cli/internal/cmds/dev/helpers"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
	"github.com/zalberix/cactus/cli/internal/cmds/dev/tui"
)

// Cmd is the top-level `dev` command.
var Cmd = &cli.Command{
	Name:  "dev",
	Usage: "Start the development environment",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "deps",
			Usage: "Run `npm install` before starting",
		},
		&cli.BoolFlag{
			Name:  "migrate",
			Usage: "Run database migrations before starting",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Build Go services without optimisations (enables dlv attach)",
		},
	},
	Subcommands: []*cli.Command{
		helpers.CleanPortsCmd,
	},
	Action: func(c *cli.Context) error {
		ctx, stop := signal.NotifyContext(c.Context, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return tui.Run(ctx, devruntime.Options{
			InstallDeps:   c.Bool("deps"),
			RunMigrations: c.Bool("migrate"),
			Debug:         c.Bool("debug"),
			LogCapacity:   2000,
		})
	},
}

func parseOnly(s string) map[string]bool {
	if s == "" {
		return nil
	}
	m := make(map[string]bool)
	for _, name := range strings.Split(s, ",") {
		if t := strings.TrimSpace(name); t != "" {
			m[t] = true
		}
	}
	return m
}

func shouldRun(name string, only map[string]bool) bool {
	if only == nil {
		return true
	}
	return only[name]
}
