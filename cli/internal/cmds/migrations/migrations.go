package migrations

import (
	"cactus/apps/core/config"
	"cactus/apps/core/pkg/db"
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
	"os"

	_ "cactus/libs/migrations/postgres"
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
		},
	}
}

func runMigration(migrationsDir string, fn func(*goose.Provider) error) error {
	cfg := config.MustLoad(nil)

	sqlxDB, err := db.New(
		context.Background(),
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.User,
		cfg.Database.Pass,
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
