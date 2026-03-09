package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.uber.org/fx"

	"github.com/jmoiron/sqlx"

	"go.temporal.io/sdk/client"
	temporalworker "github.com/zalberix/cactus/apps/core/internal/temporal/worker"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/pkg/router"
	wshub "github.com/zalberix/cactus/apps/core/internal/pkg/wshub"
	"github.com/zalberix/cactus/apps/core/internal/plugin/smtp"
	"github.com/zalberix/cactus/apps/core/internal/plugin/telegram"
	"github.com/zalberix/cactus/apps/core/internal/route"
	servermeta "github.com/zalberix/cactus/apps/core/internal/server/meta"
	coreservice "github.com/zalberix/cactus/apps/core/internal/service/core"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
	"github.com/zalberix/cactus/apps/core/internal/storage/broker"
	filestorage "github.com/zalberix/cactus/apps/core/internal/storage/file"
	pluginstorage "github.com/zalberix/cactus/apps/core/internal/storage/plugin"
	"github.com/zalberix/cactus/apps/core/internal/storage/store"
	pkgdb "github.com/zalberix/cactus/apps/core/pkg/db"
	"github.com/zalberix/cactus/apps/core/storage/db"
	"github.com/zalberix/cactus/libs/bus"
	"github.com/zalberix/cactus/libs/logger"
)

func main() {
	cfg := config.MustLoad(nil)
	slog.SetDefault(logger.SetupLogger(cfg.Env))

	app := fx.New(
		fx.Supply(cfg),
		fx.Provide(
			// Инфраструктура
			newDB,
			newBus,
			newTemporalClient,
			newTemporalWorker,
			newDBQueries,
			store.New,
			newFileStorage,
			newPluginStorage,
			newServerMeta,
			broker.New,

			// Адаптеры интерфейсов
			func(s *store.Store) coreservice.Storage { return s },
			func(s *store.Store) pipeline.Storage { return s },
			func(b *broker.Broker) coreservice.Broker { return b },
			func(f *filestorage.Storage) coreservice.FileStorage { return f },
			func(p *pluginstorage.Storage) coreservice.Plugins { return p },

			// Сервисы
			pipeline.New,
			newPipelineHub,
			coreservice.New,

			// HTTP
			newRouter,
			newHTTPServer,
		),
		fx.Invoke(
			registerPipelineHub,
			registerMetaEvent,
			registerHTTPServer,
			registerTemporalWorker,
		),
		fx.NopLogger,
	)

	app.Run()
}

func newDB(cfg *config.Config) (*sqlx.DB, error) {
	return pkgdb.New(
		context.Background(),
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.User,
		cfg.Database.Pass,
	)
}

func newBus(cfg *config.Config) (*bus.Bus, error) {
	return bus.New(cfg.Nats.URL)
}

func newDBQueries(sqlxDB *sqlx.DB) *db.Queries {
	return db.New(sqlxDB)
}

func newFileStorage() (*filestorage.Storage, error) {
	return filestorage.New("storages/local")
}

func newPluginStorage() *pluginstorage.Storage {
	s := pluginstorage.New()
	s.Add("smtp", smtp.New())
	s.Add("telegram", telegram.New())
	return s
}

func newServerMeta(cfg *config.Config) (*servermeta.ServerMeta, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить hostname: %w", err)
	}
	return &servermeta.ServerMeta{
		HostName: host,
		Port:     cfg.HTTPServer.Port,
	}, nil
}

func newPipelineHub(b *bus.Bus, svc *pipeline.Service) *wshub.PipelineHub {
	return wshub.NewPipelineHub(context.Background(), b, svc)
}

func newRouter(
	coreService *coreservice.Service,
	pipelineService *pipeline.Service,
	plugins *pluginstorage.Storage,
) *router.ServerRouter {
	r := route.New(coreService, pipelineService, plugins)
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})
	return r
}

func newHTTPServer(cfg *config.Config, r *router.ServerRouter) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%v:%v", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler:      r,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
	}
}

func registerPipelineHub(hub *wshub.PipelineHub, svc *pipeline.Service, lc fx.Lifecycle) {
	svc.SetHub(hub)
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go hub.Run()
			return nil
		},
	})
}

func registerMetaEvent(cfg *config.Config, b *broker.Broker, queries *db.Queries) error {
	ctx := context.Background()

	maxPriority, err := queries.GetMaxPriorityWeight(ctx)
	if err != nil {
		maxPriority = 0
	}

	host, _ := os.Hostname()
	endpoint := fmt.Sprintf("http://%v:%v/api/register/worker", host, cfg.HTTPServer.Port)

	return b.SendMetaEvent(ctx, maxPriority, endpoint)
}

func newTemporalClient(cfg *config.Config) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
}

func newTemporalWorker(c client.Client) temporalworker.Worker {
	return temporalworker.New(c)
}

func registerTemporalWorker(w temporalworker.Worker, c client.Client, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return w.Start()
		},
		OnStop: func(_ context.Context) error {
			w.Stop()
			c.Close()
			return nil
		},
	})
}

func registerHTTPServer(srv *http.Server, lc fx.Lifecycle, sqlxDB *sqlx.DB, b *bus.Bus) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info(
				"Запустился cactus",
				slog.String("version", "0.1.0"),
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
			defer sqlxDB.Close()
			defer b.Close()
			return srv.Shutdown(ctx)
		},
	})
}
