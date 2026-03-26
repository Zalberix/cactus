package dev

import (
	"fmt"
	"github.com/zalberix/cactus/cli/internal/cmds/dependencies"
	"os"
	"os/signal"
	"strings"
	"syscall"

	devgolang "github.com/zalberix/cactus/cli/internal/cmds/dev/golang"
	"github.com/zalberix/cactus/cli/internal/cmds/dev/helpers"
	"github.com/zalberix/cactus/cli/internal/cmds/migrations"
	"github.com/zalberix/cactus/cli/internal/cmds/proxy"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

// Cmd is the top-level `dev` command.
var Cmd = &cli.Command{
	Name:  "dev",
	Usage: "Start the development environment",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "skip-deps",
			Usage: "Skip `npm install`",
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
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		pterm.Info.Printf("Приложение запущено из: %s\n", wd)
		skipDeps := c.Bool("skip-deps")
		isDebugEnabled := c.Bool("debug")

		if isDebugEnabled {
			pterm.Info.Println("Режим отладки ВКЛЮЧЕН")
		} else {
			pterm.Info.Println("Режим отладки ВЫКЛЮЧЕН")
		}

		if err := helpers.CleanPortsCmd.Run(c); err != nil {
			return fmt.Errorf("clear ports: %w", err)
		}

		proxyStartedChan, err := proxy.StartProxy(false)
		if err != nil {
			pterm.Fatal.Printfln("Proxy failed to start: %v", err)
			return err
		}

		// wait proxy to up
		<-proxyStartedChan

		if !skipDeps {
			if err := dependencies.Cmd.Run(c); err != nil {
				return fmt.Errorf("install deps: %w", err)
			}
		}

		if err := migrations.UpMigrationCmd.Run(c); err != nil {
			pterm.Warning.Printfln("Migrations failed: %v (continuing anyway)", err)
		}

		golangApps, err := devgolang.New(isDebugEnabled)
		if err != nil {
			pterm.Fatal.Println(err)
			return err
		}

		if err := golangApps.Start(c.Context); err != nil {
			pterm.Error.Println("Error start: ", err)
			return err
		}

		exitSignal := make(chan os.Signal, 1)
		signal.Notify(exitSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		<-exitSignal
		pterm.Info.Println("Shutting down, waiting for services to stop...")
		golangApps.Stop()
		pterm.Success.Println("All services stopped")
		return nil
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
