package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalberix/cactus/apps/workers/template/config"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
	"github.com/zalberix/cactus/libs/worker"
)

func main() {
	workerIDPath := flag.String("worker-id-path", "", "path to worker ID file (.worker_id/{name}/{uuid})")
	workerVariant := flag.String("worker-variant", "html", "worker manifest and handler variant")
	workerName := flag.String("worker-name", "", "runtime worker name for registration")
	flag.Parse()

	cfg := cfgloader.MustLoad[config.Config]("configs/workers/template.yaml")
	slog.SetDefault(logger.SetupLogger(cfg.Env, logger.WorkerSource("template", 0)))
	log := slog.Default()

	selectedVariant, err := worker.SelectVariant(templateVariants(), *workerVariant)
	if err != nil {
		slog.Error("select worker variant", slog.String("variant", *workerVariant), slog.String("error", err.Error()))
		os.Exit(1)
	}

	w := worker.New(worker.Config{
		NatsURL:        cfg.NatsURL,
		ManagerURL:     cfg.ManagerURL,
		BootstrapToken: cfg.BootstrapToken,
		WorkerIDPath:   *workerIDPath,
		WorkerName:     templateWorkerName(selectedVariant.Name, *workerName),
		OnWorkerID: func(workerID int32) *slog.Logger {
			return logger.SetupLogger(cfg.Env, templateWorkerSource(selectedVariant.Name, workerID))
		},
		Manifest: selectedVariant.Manifest,
	}, selectedVariant.Handler, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("Template worker starting", slog.String("variant", selectedVariant.Name))

	if err := w.Run(ctx); err != nil {
		slog.Error("worker stopped", slog.String("error", err.Error()))
		cancel()
		return
	}

	slog.Info("Template worker stopped")
}

func templateWorkerSource(variant string, workerID int32) string {
	return logger.WorkerSource("template-"+variant, workerID)
}

func templateWorkerName(variant, override string) string {
	if override != "" {
		return override
	}
	return "template-" + variant + "-worker"
}
