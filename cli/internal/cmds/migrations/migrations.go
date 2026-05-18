package migrations

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/zalberix/cactus/apps/core/config"
	pkgdb "github.com/zalberix/cactus/apps/core/pkg/db"
	cfgloader "github.com/zalberix/cactus/libs/config"

	// Register embedded postgres migrations for goose.
	_ "github.com/zalberix/cactus/libs/migrations/postgres"
)

// const migrationsDir = "./libs/migrations/postgres"

func Command() *cli.Command {
	return &cli.Command{
		Name:    "migration",
		Aliases: []string{"migrations"},
		Usage:   "Database migration commands",
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
	cfg := cfgloader.MustLoad[config.Config]("configs/apps/core.yaml")

	pool, err := pkgdb.New(
		context.Background(),
		cfg.Database.URL,
	)
	if err != nil {
		pterm.Error.Printfln("DB connection failed: %v", err)
		return err
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, os.DirFS(migrationsDir))
	if err != nil {
		pterm.Error.Printfln("Failed to create goose provider: %v", err)
		return err
	}

	return fn(provider)
}

func runMigrationWithPool(fn func(*pgxpool.Pool) error) error {
	cfg := cfgloader.MustLoad[config.Config]("configs/apps/core.yaml")

	pool, err := pkgdb.New(
		context.Background(),
		cfg.Database.URL,
	)
	if err != nil {
		pterm.Error.Printfln("DB connection failed: %v", err)
		return err
	}
	defer pool.Close()

	return fn(pool)
}
