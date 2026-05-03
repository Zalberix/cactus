package migrations

import (
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var DownMigrationCmd = &cli.Command{
	Name:  "down",
	Usage: "Roll back the last applied migration",
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

		// Пока только postgres, других миграций нет
		migrationsPath := filepath.Join(wd, c.String("migrations-path"), "postgres")

		err = runMigration(migrationsPath, func(p *goose.Provider) error {
			result, err := p.Down(c.Context)
			if err != nil {
				return err
			}
			pterm.Success.Printfln("Rolled back: %s", result.Source.Path)
			return nil
		})
		if err != nil {
			return err
		}

		pterm.Success.Println("Migration succeed")
		return nil
	},
}
