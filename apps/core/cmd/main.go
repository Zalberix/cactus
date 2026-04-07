package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/domain/auth"
	"github.com/zalberix/cactus/apps/core/internal/domain/message"
	"github.com/zalberix/cactus/apps/core/internal/domain/rbac"
	"github.com/zalberix/cactus/apps/core/internal/domain/workflow"
	"github.com/zalberix/cactus/apps/core/internal/domain/worktype"
	apphttp "github.com/zalberix/cactus/apps/core/internal/http"
	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
	"github.com/zalberix/cactus/apps/core/internal/pkg/wshub"
	"github.com/zalberix/cactus/apps/core/internal/store"
	temporalactivity "github.com/zalberix/cactus/apps/core/internal/temporal/activity"
	temporalworker "github.com/zalberix/cactus/apps/core/internal/temporal/worker"
	pkgdb "github.com/zalberix/cactus/apps/core/pkg/db"
	"github.com/zalberix/cactus/libs/bus"
	cfgloader "github.com/zalberix/cactus/libs/config"
	"github.com/zalberix/cactus/libs/logger"
)

func main() {
	cfg := cfgloader.MustLoad[config.Config]("configs/apps/core.yaml")
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
			newRBACService,
			rbac.NewHandler,
			newWorktypeService,
			newWorktypeHandler,
			newWorkflowService,
			newWorkflowHandler,
			temporalworker.NewTemporalClient,
			temporalactivity.New,
			newMessageService,
			newMessageHandler,
			newWSHub,
		),
		fx.Invoke(
			registerNATSStreams,
			registerAuthRoutes,
			registerRBACRoutes,
			registerWorkTypeRoutes,
			registerWorkflowRoutes,
			registerMessageRoutes,
			registerWSRoutes,
			temporalworker.RegisterTemporalWorker,
			registerHTTPServer,
		),
		//fx.NopLogger,
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
			// Старые streams (обратная совместимость, Phase 4 мигрирует воркеров)
			if err := b.EnsureStream(ctx, "MESSAGES", []string{"messages.>"}); err != nil {
				return fmt.Errorf("ensure MESSAGES stream: %w", err)
			}
			if err := b.EnsureStream(ctx, "EVENTS", []string{"event.>"}); err != nil {
				return fmt.Errorf("ensure EVENTS stream: %w", err)
			}
			if err := b.EnsureStream(ctx, "TASKS", []string{"task.>"}); err != nil {
				return fmt.Errorf("ensure TASKS stream: %w", err)
			}
			if err := b.EnsureStreamWithMaxAge(ctx, "RESULTS", []string{"result.>"}, time.Hour); err != nil {
				return fmt.Errorf("ensure RESULTS stream: %w", err)
			}
			slog.Info("NATS JetStream streams ensured", slog.String("streams", "MESSAGES, EVENTS, TASKS, RESULTS"))
			return nil
		},
	})
}

// newRBACService создаёт rbac.Service, передавая store как Storage и authService как AuthService.
func newRBACService(s *store.Store, authSvc *auth.Service) *rbac.Service {
	return rbac.NewService(s, authSvc)
}

// registerRBACRoutes регистрирует защищённые RBAC маршруты.
func registerRBACRoutes(r *gin.Engine, h *rbac.Handler, authSvc *auth.Service, s *store.Store) {
	authMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, authMw, s)
}

// newWorktypeService создаёт worktype.Service, передавая store как Storage.
func newWorktypeService(s *store.Store) *worktype.Service {
	return worktype.NewService(s)
}

// newWorktypeHandler создаёт worktype.Handler.
func newWorktypeHandler(svc *worktype.Service, s *store.Store) *worktype.Handler {
	return worktype.NewHandler(svc, s)
}

// registerWorkTypeRoutes регистрирует маршруты worktype домена.
func registerWorkTypeRoutes(r *gin.Engine, h *worktype.Handler, authSvc *auth.Service, s *store.Store) {
	authMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, authMw, s)
}

// newWorkflowService создаёт workflow.Service, передавая store как Storage.
func newWorkflowService(s *store.Store) *workflow.Service {
	return workflow.NewService(s)
}

// newWorkflowHandler создаёт workflow.Handler.
func newWorkflowHandler(svc *workflow.Service, s *store.Store) *workflow.Handler {
	return workflow.NewHandler(svc, s)
}

// registerWorkflowRoutes регистрирует маршруты workflow домена.
func registerWorkflowRoutes(r *gin.Engine, h *workflow.Handler, authSvc *auth.Service) {
	authMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, authMw)
}

// newMessageService создаёт message.Service с Temporal client.
func newMessageService(s *store.Store, tc client.Client) *message.Service {
	return message.NewService(s, tc)
}

// newMessageHandler создаёт message.Handler.
func newMessageHandler(svc *message.Service) *message.Handler {
	return message.NewHandler(svc)
}

// registerMessageRoutes регистрирует маршруты message domain.
// Per D-11: POST /api/v1/messages/send доступен через system token auth (M2M)
// и POST /api/v1/messages/send-user через JWT auth (UI).
func registerMessageRoutes(r *gin.Engine, h *message.Handler, s *store.Store, authSvc *auth.Service) {
	systemTokenMw := middleware.SystemTokenAuth(s)
	jwtMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, jwtMw, systemTokenMw)
}

// newWSHub creates the WebSocket Hub for real-time workflow status.
func newWSHub(b *bus.Bus, msgSvc *message.Service, authSvc *auth.Service, s *store.Store) *wshub.Hub {
	return wshub.New(b, msgSvc, authSvc, s)
}

// registerWSRoutes registers the WebSocket endpoint
// No auth middleware here -- auth happens inside the WS handshake
func registerWSRoutes(r *gin.Engine, hub *wshub.Hub) {
	r.GET("/ws/workflow/:messageID", hub.HandleWS)
}

func registerHTTPServer(srv *http.Server, lc fx.Lifecycle, pool *pgxpool.Pool, b *bus.Bus, tc client.Client) {
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
			tc.Close() // Close Temporal client
			return srv.Shutdown(ctx)
		},
	})
}
