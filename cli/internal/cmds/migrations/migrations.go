package migrations

import (
	"context"
	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/pkg/db"
	"os"

	_ "github.com/zalberix/cactus/libs/migrations/postgres"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

//const migrationsDir = "./libs/migrations/postgres"

func Command() *cli.Command {
	return &cli.Command{
		Name:  "migration",
		Usage: "Database migration commands",
		Subcommands: []*cli.Command{
			UpMigrationCmd,
			DownMigrationCmd,
			StatusMigrationCmd,
			CreateMigrationCmd,
			ResetMigrationCmd,
		},
	}
}

func runMigration(migrationsDir string, fn func(*goose.Provider) error) error {
	cfg := config.MustLoad(nil)

	sqlxDB, err := db.New(
		context.Background(),
		cfg.Database.URL,
	)
	if err != nil {
		pterm.Error.Printfln("DB connection failed: %v", err)
		return err
	}
	defer func(sqlxDB *sqlx.DB) {
		err := sqlxDB.Close()
		if err != nil {
			pterm.Error.Printfln("Failed close db connection: %v", err)
			return
		}
	}(sqlxDB)

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlxDB.DB, os.DirFS(migrationsDir))
	if err != nil {
		pterm.Error.Printfln("Failed to create goose provider: %v", err)
		return err
	}

	return fn(provider)
}
