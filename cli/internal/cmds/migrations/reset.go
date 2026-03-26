package migrations

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var ResetMigrationCmd = &cli.Command{
	Name:  "reset",
	Usage: "Drop all migrations and re-apply (dev only)",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "migrations-path",
			Value: "./libs/migrations",
		},
	},
	Action: func(c *cli.Context) error {
		pterm.Warning.Println("Resetting database (drop all + re-apply)...")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		migrationsPath := filepath.Join(wd, c.String("migrations-path"), "postgres")

		return runMigration(migrationsPath, func(provider *goose.Provider) error {
			for {
				result, err := provider.Down(c.Context)
				if err != nil {
					if errors.Is(err, goose.ErrNoNextVersion) {
						break
					}
					return err
				}
				if result.Source.Path == "" {
					break
				}
				pterm.Info.Printfln("Rolled back: %s", result.Source.Path)
			}

			results, err := provider.Up(c.Context)
			if err != nil {
				return err
			}
			for _, r := range results {
				pterm.Success.Printfln("Applied: %s", r.Source.Path)
			}

			pterm.Success.Println("Database reset complete")
			return nil
		})
	},
}
