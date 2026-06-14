package app

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

	"github.com/zalberix/cactus/apps/manager/config"
	"github.com/zalberix/cactus/apps/manager/internal/configpub"
	"github.com/zalberix/cactus/apps/manager/internal/domain/auth"
	"github.com/zalberix/cactus/apps/manager/internal/domain/message"
	"github.com/zalberix/cactus/apps/manager/internal/domain/rbac"
	"github.com/zalberix/cactus/apps/manager/internal/domain/workflow"
	"github.com/zalberix/cactus/apps/manager/internal/domain/worktype"
	apphttp "github.com/zalberix/cactus/apps/manager/internal/http"
	"github.com/zalberix/cactus/apps/manager/internal/http/middleware"
	"github.com/zalberix/cactus/apps/manager/internal/natsauth"
	"github.com/zalberix/cactus/apps/manager/internal/pkg/wshub"
	"github.com/zalberix/cactus/apps/manager/internal/store"
	pkgdb "github.com/zalberix/cactus/apps/manager/pkg/db"
	"github.com/zalberix/cactus/libs/bus"
)

func NewManager(cfg *config.Config) *fx.App {
	return fx.New(ManagerOptions(cfg))
}

func ManagerOptions(cfg *config.Config) fx.Option {
	return fx.Options(
		fx.Supply(cfg),
		fx.Provide(managerProviders()...),
		fx.Invoke(managerInvokes()...),
	)
}

func commonProviders() []any {
	return []any{
		pkgdb.NewFx,
		store.New,
		newBus,
		newConfigPubService,
		newTemporalClient,
	}
}

func managerProviders() []any {
	providers := append([]any{}, commonProviders()...)
	providers = append(providers,
		apphttp.NewRouter,
		apphttp.NewHTTPServer,
		apphttp.NewDebugServer,
		newAuthService,
		auth.NewHandler,
		newRBACService,
		rbac.NewHandler,
		newNATSAuthManager,
		newWorktypeService,
		newWorktypeHandler,
		newWorkflowService,
		newWorkflowHandler,
		newMessageService,
		newMessageHandler,
		newWSHub,
	)
	return providers
}

func managerInvokes() []any {
	return []any{
		registerNATSStreams,
		registerConfigRequestListener,
		registerResourceCloser,
		registerAuthRoutes,
		registerRBACRoutes,
		registerWorkTypeRoutes,
		registerWorkflowRoutes,
		registerMessageRoutes,
		registerWSRoutes,
		registerDebugServer,
		registerHTTPServer,
	}
}

func newAuthService(cfg *config.Config, s *store.Store) *auth.Service {
	return auth.NewService(cfg.JWT, s)
}

func newBus(cfg *config.Config) (*bus.Bus, error) {
	return bus.NewWithOptions(bus.Options{
		URL:             cfg.Nats.URL,
		CAFile:          cfg.Nats.CAFile,
		CredentialsFile: cfg.Nats.CredentialsFile,
	})
}

func newTemporalClient(cfg *config.Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal client dial: %w", err)
	}
	slog.Info("Temporal client connected",
		slog.String("host", cfg.Temporal.HostPort),
		slog.String("namespace", cfg.Temporal.Namespace),
	)
	return c, nil
}

func registerAuthRoutes(r *gin.Engine, h *auth.Handler, authSvc *auth.Service) {
	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/refresh", h.Refresh)
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
			if err := b.EnsureStream(ctx, "TASKS", []string{"task.>"}); err != nil {
				return fmt.Errorf("ensure TASKS stream: %w", err)
			}
			if err := b.EnsureStreamWithMaxAge(ctx, "RESULTS", []string{"result.>"}, time.Hour); err != nil {
				return fmt.Errorf("ensure RESULTS stream: %w", err)
			}
			if err := b.EnsureStream(ctx, "CONFIGS", []string{"config.org.*.work_type.*.revision.*"}, bus.WithAllowDirect()); err != nil {
				return fmt.Errorf("ensure CONFIGS stream: %w", err)
			}
			slog.Info("NATS JetStream streams ensured", slog.String("streams", "MESSAGES, EVENTS, TASKS, RESULTS, CONFIGS"))
			return nil
		},
	})
}

func newConfigPubService(s *store.Store, b *bus.Bus) *configpub.Service {
	return configpub.New(s, b)
}

func registerConfigRequestListener(lc fx.Lifecycle, svc *configpub.Service) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return svc.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return svc.Stop(ctx)
		},
	})
}

func newRBACService(s *store.Store, authSvc *auth.Service) *rbac.Service {
	return rbac.NewService(s, authSvc)
}

func registerRBACRoutes(r *gin.Engine, h *rbac.Handler, authSvc *auth.Service, s *store.Store) {
	authMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, authMw, s)
}

func newNATSAuthManager(lc fx.Lifecycle, cfg *config.Config) natsauth.Manager {
	if !cfg.Nats.WorkerCredentialsEnabled {
		return natsauth.NoopManager{}
	}
	updater, err := natsauth.NewResolverUpdater(
		cfg.Nats.URL,
		cfg.Nats.CAFile,
		cfg.Nats.SystemCredentialsFile,
	)
	if err != nil {
		panic(err)
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return updater.Close()
		},
	})
	return natsauth.NewJWTManager(
		cfg.Nats.AccountPublicKey,
		cfg.Nats.AccountJWTFile,
		cfg.Nats.AccountSeedEnv,
		cfg.Nats.AccountSeedFile,
		updater,
		natsauth.WithOperatorSeedFile(cfg.Nats.OperatorSeedFile),
	)
}

func newWorktypeService(cfg *config.Config, s *store.Store, auth natsauth.Manager) *worktype.Service {
	return worktype.NewService(
		s,
		auth,
		worktype.WithNATSURL(cfg.Nats.URL),
		worktype.WithNATSCAFile(cfg.Nats.CAFile),
	)
}

func newWorktypeHandler(svc *worktype.Service, s *store.Store) *worktype.Handler {
	return worktype.NewHandler(svc, s)
}

func registerWorkTypeRoutes(r *gin.Engine, h *worktype.Handler, authSvc *auth.Service, s *store.Store) {
	authMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, authMw, s)
}

func newWorkflowService(s *store.Store) *workflow.Service {
	return workflow.NewService(s)
}

func newWorkflowHandler(svc *workflow.Service, s *store.Store) *workflow.Handler {
	return workflow.NewHandler(svc, s)
}

func registerWorkflowRoutes(r *gin.Engine, h *workflow.Handler, authSvc *auth.Service) {
	authMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, authMw)
}

func newMessageService(s *store.Store, tc client.Client) *message.Service {
	return message.NewService(s, tc)
}

func newMessageHandler(svc *message.Service) *message.Handler {
	return message.NewHandler(svc)
}

func registerMessageRoutes(r *gin.Engine, h *message.Handler, s *store.Store, authSvc *auth.Service) {
	systemTokenMw := middleware.SystemTokenAuth(s)
	jwtMw := middleware.Auth(authSvc)
	h.RegisterRoutes(r, jwtMw, systemTokenMw)
}

func newWSHub(b *bus.Bus, msgSvc *message.Service, authSvc *auth.Service, s *store.Store) *wshub.Hub {
	return wshub.New(b, msgSvc, authSvc, s)
}

func registerWSRoutes(r *gin.Engine, hub *wshub.Hub) {
	r.GET("/ws/workflow/:messageID", hub.HandleWS)
}

func registerDebugServer(debug *apphttp.DebugServer, lc fx.Lifecycle) {
	if debug == nil || !debug.Enabled || debug.Server == nil {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info("Core pprof debug server started", slog.String("addr", debug.Server.Addr))
			go func() {
				if err := debug.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("Core pprof debug server error", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Core pprof debug server stopping")
			return debug.Server.Shutdown(ctx)
		},
	})
}

func registerResourceCloser(lc fx.Lifecycle, pool *pgxpool.Pool, b *bus.Bus, tc client.Client) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			pool.Close()
			b.Close()
			tc.Close()
			return nil
		},
	})
}

func registerHTTPServer(srv *http.Server, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info(
				"Started cactus",
				slog.String("version", "0.2.0"),
				slog.String("addr", srv.Addr),
			)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("Server start error", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Stopping server")
			return srv.Shutdown(ctx)
		},
	})
}
