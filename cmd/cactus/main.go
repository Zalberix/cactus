package main

import (
	"cactus/internal/config"
	"cactus/internal/logger"
	"cactus/internal/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cfg := config.MustLoad()

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

	if err := serverApp.Start(); err != nil {
		slog.Error("Ошибка запуска сервера")
	}

	slog.Info("Сервер запущен")
	<-sigTerm
	slog.Info("Остановка сервера")
	// тут надо сделать сохранение состояния и очистку памяти и тд
}
