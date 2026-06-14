package seed

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/zalberix/cactus/apps/manager/config"
	pkgdb "github.com/zalberix/cactus/apps/manager/pkg/db"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/migrations/seeds"
)

var Cmd = &cli.Command{
	Name:  "seed",
	Usage: "Insert seed data (test org, admin user, permissions, work types)",
	Action: func(c *cli.Context) error {
		pterm.Info.Println("Seeding database...")

		cfg := cfgloader.MustLoad[config.Config]("configs/apps/manager.yaml")
		db, err := pkgdb.New(context.Background(), cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("db connection failed: %w", err)
		}
		defer db.Close()

		if err := seeds.SeedAll(c.Context, db); err != nil {
			pterm.Error.Printfln("Seed failed: %v", err)
			return err
		}

		pterm.Success.Println("Seed complete")
		return nil
	},
}
