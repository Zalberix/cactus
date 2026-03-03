package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cactus/apps/core/config"
	"cactus/apps/core/internal/server"
	"cactus/libs/shared/logger"
)

func main() {
	cfg := config.MustLoad(nil)

	// isDev := cfg.Env == config.AppEnvDevelopment || cfg.Env == config.AppEnvLocal

	slog.SetDefault(logger.SetupLogger(cfg.Env))

	serverApp, err := server.Create(*cfg)
	if err != nil {
		slog.Error(
			"Инициализация сервера",
			slog.Any("error", err.Error()),
		)
		return
	}

	slog.Info(
		"Запустился cactus",
		slog.String("version", "0.1.0"),
	)

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := serverApp.Start(); err != nil {
			slog.Error("Ошибка запуска сервера", slog.String("error", err.Error()))
			sigTerm <- os.Interrupt
		}
	}()

	slog.Info("Сервер запущен")
	<-sigTerm
	slog.Info("Остановка сервера")
	// TODO тут надо сделать сохранение состояния и очистку памяти и тд
}
