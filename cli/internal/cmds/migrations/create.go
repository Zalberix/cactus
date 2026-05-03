package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var CreateMigrationCmd = &cli.Command{
	Name:  "create",
	Usage: "Create a new migration file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Usage:    "Name of the migration",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "type",
			Usage:    "Migration type (sql, go)",
			Required: false,
		},
	},
	ArgsUsage: "<name> [sql|go]",
	Action: func(c *cli.Context) error {
		name, db, migrationType := c.String("name"), "postgres", c.String("type")

		pterm.Info.Println("Create migrations")

		if name == "" {
			migrationNameInput := pterm.DefaultInteractiveTextInput.WithMultiLine(false)
			migrationName, err := migrationNameInput.Show("Migration name")
			if err != nil {
				return err
			}

			name = strings.TrimSpace(migrationName)
		}

		if migrationType == "" {
			mType, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"sql", "go"}).Show()
			if err != nil {
				return err
			}

			migrationType = mType
		}

		if migrationType != "sql" && migrationType != "go" {
			return fmt.Errorf("invalid migration type: %s", migrationType)
		}

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		dir := filepath.Join(wd, "libs", "migrations", db)

		if err := goose.Create(nil, dir, name, migrationType); err != nil {
			return err
		}
		// fix package name in go migrations
		if migrationType == "go" {
			files, err := os.ReadDir(dir)
			if err != nil {
				return fmt.Errorf("cannot read migrations directory: %w", err)
			}

			var latestMigrationFile string
			for _, file := range files {
				if file.IsDir() || filepath.Ext(file.Name()) != ".go" {
					continue
				}

				latestMigrationFile = file.Name()
			}

			if latestMigrationFile != "" {
				migrationFilePath := filepath.Join(dir, latestMigrationFile)
				content, err := os.ReadFile(migrationFilePath)
				if err != nil {
					return fmt.Errorf("cannot read migration file: %w", err)
				}

				fixedContent := strings.Replace(
					string(content),
					"package migrations",
					fmt.Sprintf("package %s", db),
					1,
				)
				err = os.WriteFile(migrationFilePath, []byte(fixedContent), 0o600)
				if err != nil {
					return fmt.Errorf("cannot write fixed migration file: %w", err)
				}

				pterm.Info.Println("Fixed package name in migration file")
			}
		}

		pterm.Success.Printfln("Created %s migration: %s", migrationType, name)
		return nil
	},
}
