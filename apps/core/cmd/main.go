package main

import (
	"log/slog"

	"github.com/zalberix/cactus/apps/core/config"
	coreapp "github.com/zalberix/cactus/apps/core/internal/app"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
)

func main() {
	cfg := cfgloader.MustLoad[config.Config]("configs/apps/core.yaml")
	slog.SetDefault(logger.SetupLogger(cfg.Env, "core-worker"))

	coreapp.NewWorker(cfg).Run()
}
