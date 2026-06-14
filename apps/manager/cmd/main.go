package main

import (
	"log/slog"

	"github.com/zalberix/cactus/apps/manager/config"
	managerapp "github.com/zalberix/cactus/apps/manager/internal/app"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
)

func main() {
	cfg := cfgloader.MustLoad[config.Config]("configs/apps/manager.yaml")
	slog.SetDefault(logger.SetupLogger(cfg.Env, "manager"))

	managerapp.NewManager(cfg).Run()
}
