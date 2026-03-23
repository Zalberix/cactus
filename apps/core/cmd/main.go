package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	apphttp "github.com/zalberix/cactus/apps/core/internal/http"
	"github.com/zalberix/cactus/apps/core/internal/store"
	pkgdb "github.com/zalberix/cactus/apps/core/pkg/db"
	"github.com/zalberix/cactus/libs/bus"
	"github.com/zalberix/cactus/libs/logger"
)

func main() {
	cfg := config.MustLoad(nil)
	slog.SetDefault(logger.SetupLogger(cfg.Env))

	app := fx.New(
		fx.Supply(cfg),
		fx.Provide(
			pkgdb.NewFx,
			store.New,
			bus.NewFx,
			apphttp.NewRouter,
			apphttp.NewHTTPServer,
		),
		fx.Invoke(
			registerNATSStreams,
			registerHTTPServer,
		),
		fx.NopLogger,
	)

	app.Run()
}

func registerNATSStreams(lc fx.Lifecycle, b *bus.Bus) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := b.EnsureStream(ctx, "MESSAGES", []string{"messages.>"}); err != nil {
				return fmt.Errorf("ensure MESSAGES stream: %w", err)
			}
			if err := b.EnsureStream(ctx, "EVENTS", []string{"event.>"}); err != nil {
				return fmt.Errorf("ensure EVENTS stream: %w", err)
			}
			slog.Info("NATS JetStream streams ensured", slog.String("streams", "MESSAGES, EVENTS"))
			return nil
		},
	})
}

func registerHTTPServer(srv *http.Server, lc fx.Lifecycle, pool *pgxpool.Pool, b *bus.Bus) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info(
				"Запустился cactus",
				slog.String("version", "0.2.0"),
				slog.String("addr", srv.Addr),
			)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("Ошибка запуска сервера", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Остановка сервера")
			pool.Close()
			b.Close()
			return srv.Shutdown(ctx)
		},
	})
}
