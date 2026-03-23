package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/domain/auth"
	apphttp "github.com/zalberix/cactus/apps/core/internal/http"
	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
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
			newAuthService,
			auth.NewHandler,
		),
		fx.Invoke(
			registerNATSStreams,
			registerAuthRoutes,
			registerHTTPServer,
		),
		fx.NopLogger,
	)

	app.Run()
}

// newAuthService создаёт auth.Service, передавая JWT конфиг и store как Storage.
func newAuthService(cfg *config.Config, s *store.Store) *auth.Service {
	return auth.NewService(cfg.JWT, s)
}

// registerAuthRoutes регистрирует публичные маршруты аутентификации
// и создаёт защищённую группу /api/v1 с JWT middleware для планов 02-03..02-05.
func registerAuthRoutes(r *gin.Engine, h *auth.Handler, authSvc *auth.Service) {
	v1 := r.Group("/api/v1")

	// Публичные маршруты
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/refresh", h.Refresh)

	// Защищённая группа — маршруты планов 02-03, 02-04, 02-05 регистрируются в registerProtectedRoutes
	_ = v1.Group("", middleware.Auth(authSvc))
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
