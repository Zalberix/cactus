package migrations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var StatusMigrationCmd = &cli.Command{
	Name:  "status",
	Usage: "Show migration status",
	Action: func(c *cli.Context) error {
		pterm.Info.Println("Running migrations")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Пока только postgres, других миграций нет
		migrationsPath := filepath.Join(wd, c.String("migrations-path"), "postgres")

		err = runMigration(migrationsPath, func(p *goose.Provider) error {
			statuses, err := p.Status(c.Context)
			if err != nil {
				return err
			}

			td := pterm.TableData{{"Version", "Applied", "Name"}}
			for _, s := range statuses {
				applied := "pending"
				if s.State == goose.StateApplied {
					applied = "applied"
				}
				td = append(td, []string{
					fmt.Sprintf("%d", s.Source.Version),
					applied,
					s.Source.Path,
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithData(td).Render()
		})
		if err != nil {
			return err
		}

		pterm.Success.Println("Migration succeed")
		return nil
	},
}
