package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/configpub"
	"github.com/zalberix/cactus/apps/core/internal/store"
	temporalactivity "github.com/zalberix/cactus/apps/core/internal/temporal/activity"
	temporalworker "github.com/zalberix/cactus/apps/core/internal/temporal/worker"
	pkgdb "github.com/zalberix/cactus/apps/core/pkg/db"
	"github.com/zalberix/cactus/libs/bus"
)

func NewWorker(cfg *config.Config) *fx.App {
	return fx.New(WorkerOptions(cfg))
}

func WorkerOptions(cfg *config.Config) fx.Option {
	return fx.Options(
		fx.Supply(cfg),
		fx.Provide(workerProviders()...),
		fx.Invoke(workerInvokes()...),
	)
}

func workerProviders() []any {
	return []any{
		pkgdb.NewFx,
		store.New,
		newBus,
		newConfigPubService,
		temporalworker.NewTemporalClient,
		temporalactivity.New,
	}
}

func workerInvokes() []any {
	return []any{
		registerNATSStreams,
		registerResourceCloser,
		temporalworker.RegisterTemporalWorker,
	}
}

func newBus(cfg *config.Config) (*bus.Bus, error) {
	return bus.NewWithOptions(bus.Options{
		URL:             cfg.Nats.URL,
		CAFile:          cfg.Nats.CAFile,
		CredentialsFile: cfg.Nats.CredentialsFile,
	})
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
