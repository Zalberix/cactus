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
	"github.com/zalberix/cactus/apps/core/internal/temporal"
	temporalworker "github.com/zalberix/cactus/apps/core/internal/temporal/worker"
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
			pkgdb.NewFx,
			bus.NewFx,
			temporal.NewClientFx,
			temporalworker.NewFx,
			db.NewFx,
			store.NewFx,
			filestorage.NewFx,
			smtp.NewFx,
			telegram.NewFx,
			pluginstorage.NewFx,
			servermeta.NewFx,
			broker.NewFx,

			// Адаптеры интерфейсов
			func(s *store.Store) coreservice.Storage { return s },
			func(s *store.Store) pipeline.Storage { return s },
			func(b *broker.Broker) coreservice.Broker { return b },
			func(f *filestorage.Storage) coreservice.FileStorage { return f },
			func(p *pluginstorage.Storage) coreservice.Plugins { return p },
			func(s *pipeline.Service) wshub.ServicePipelineHub { return s },

			// Сервисы
			pipeline.NewFx,
			wshub.NewPipelineHubFx,
			coreservice.NewFx,

			// HTTP
			route.NewFx,
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

func newHTTPServer(cfg *config.Config, r *router.ServerRouter) *http.Server {
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

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
