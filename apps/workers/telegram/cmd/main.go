package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalberix/cactus/apps/workers/telegram/config"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
	"github.com/zalberix/cactus/libs/worker"
)

func main() {
	workerIDPath := flag.String("worker-id-path", "", "path to worker ID file (.worker_id/{type}/{uuid})")
	workerVariant := flag.String("worker-variant", "basic", "worker manifest and handler variant")
	workerName := flag.String("worker-name", "", "runtime worker name for registration")
	flag.Parse()

	conf := cfgloader.MustLoad[config.Config]("configs/apps/workers/telegram.yaml")
	slog.SetDefault(logger.SetupLogger(conf.Env, logger.WorkerSource("telegram", 0)))

	bootstrapToken := conf.BootstrapToken
	if bootstrapToken == "" {
		bootstrapToken = conf.Token
	}
	if bootstrapToken == "" {
		slog.Error("bootstrap token is required")
		os.Exit(1)
	}

	selectedVariant, err := worker.SelectVariant(telegramVariants(*conf), *workerVariant)
	if err != nil {
		slog.Error("select worker variant", slog.String("variant", *workerVariant), slog.String("error", err.Error()))
		os.Exit(1)
	}

	workerCore := worker.New(worker.Config{
		NatsURL:        conf.NatsURL,
		ManagerURL:     conf.ManagerURL,
		BootstrapToken: bootstrapToken,
		WorkerIDPath:   *workerIDPath,
		WorkerName:     telegramWorkerName(selectedVariant.Name, *workerName),
		OnWorkerID: func(workerID int32) *slog.Logger {
			return logger.SetupLogger(conf.Env, telegramWorkerSource(selectedVariant.Name, workerID))
		},
		Manifest: selectedVariant.Manifest,
	}, selectedVariant.Handler, slog.Default())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("Telegram worker starting", slog.String("variant", selectedVariant.Name))

	if err := workerCore.Run(ctx); err != nil {
		slog.Error("worker stopped", slog.String("error", err.Error()))
		cancel()
		return
	}
}

func telegramWorkerSource(variant string, workerID int32) string {
	return logger.WorkerSource("telegram-"+variant, workerID)
}

func telegramWorkerName(variant, override string) string {
	if override != "" {
		return override
	}
	return "telegram-" + variant + "-worker"
}
