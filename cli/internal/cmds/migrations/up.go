package migrations

import (
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var UpMigrationCmd = &cli.Command{
	Name:  "up",
	Usage: "Apply all pending migrations",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "migrations-path",
			Value: "./libs/migrations",
		},
	},
	Action: func(c *cli.Context) error {
		pterm.Info.Println("Running migrations")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		pterm.Info.Println(wd)

		// Пока только postgres, других миграций нет
		migrationsPath := filepath.Join(wd, c.String("migrations-path"), "postgres")

		err = runMigration(migrationsPath, func(provider *goose.Provider) error {
			result, err := provider.Up(c.Context)
			if err != nil {
				return err
			}
			if len(result) == 0 {
				pterm.Info.Println("No pending migrations")
				return nil
			}
			for _, r := range result {
				pterm.Success.Printfln("Applied: %s", r.Source.Path)
			}

			return nil
		})
		if err != nil {
			return err
		}

		pterm.Success.Println("Migration succeed")
		return nil
	},
}
