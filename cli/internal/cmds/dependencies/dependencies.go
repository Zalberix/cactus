package dependencies

import (
	"cactus/cli/internal/shell"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/pterm/pterm"
)

// Install runs `npm install` at the monorepo root.
func Install(rootDir string) error {
	pterm.Info.Println("Installing dependencies...")
	return shell.ExecCommand(
		shell.ExecCommandOpts{
			Command: "npm install",
			Pwd:     rootDir,
		},
	)
}

var Cmd = &cli.Command{
	Name:    "dependencies",
	Usage:   "install golang and nodejs dependencies",
	Aliases: []string{"deps"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "skip-node",
			Value: false,
			Usage: "skip nodejs dependencies installation",
		},
		&cli.BoolFlag{
			Name:  "skip-go",
			Value: false,
			Usage: "skip golang dependencies installation",
		},
	},
	Action: func(c *cli.Context) error {
		skipNode := c.Bool("skip-node")
		skipGo := c.Bool("skip-go")

		var wg errgroup.Group
		if !skipGo {
			wg.Go(
				func() error {
					if err := installGolangDeps(); err != nil {
						pterm.Fatal.Println(err)
						return err
					}
					pterm.Success.Println("Golang deps installed")
					return nil
				},
			)
		} else {
			go pterm.Warning.Println("Golang deps skipped")
		}

		if !skipNode {
			wg.Go(
				func() error {
					if err := installNodeDeps(); err != nil {
						pterm.Fatal.Println(err)
						return err
					}
					pterm.Success.Println("Nodejs deps installed")
					return nil
				},
			)
		} else {
			go pterm.Warning.Println("Nodejs deps skipped")
		}

		if err := wg.Wait(); err != nil {
			return err
		}

		pterm.Success.Println("All dependencies installed")

		return nil
	},
}
