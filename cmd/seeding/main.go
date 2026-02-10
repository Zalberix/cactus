package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cactus/cmd/seeding/seeds"
	"cactus/internal/config"
	"cactus/internal/logger"
	sqlxconect "cactus/internal/pkg/db"
	"cactus/internal/storage/db"
)

func main() {
	cfg := config.MustLoad()
	slog.SetDefault(logger.SetupLogger(cfg.Env))

	seedName := "all"
	if len(os.Args) > 1 {
		seedName = os.Args[1]
	}

	ctx := context.Background()

	databaseConect, err := sqlxconect.New(
		ctx,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.User,
		cfg.Database.Pass,
	)
	if err != nil {
		slog.Error("ошибка подключения к БД", slog.String("error", err.Error()))
		os.Exit(1)
	}

	storage := db.New(databaseConect)

	if seedName == "list" {
		fmt.Println("Доступные сиды:")
		for _, name := range seeds.List() {
			fmt.Printf("  - %s\n", name)
		}
		return
	}

	if err := seeds.Run(ctx, storage, seedName); err != nil {
		slog.Error("ошибка выполнения сида", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
