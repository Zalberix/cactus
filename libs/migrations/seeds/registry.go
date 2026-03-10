package seeds

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zalberix/cactus/apps/core/storage/db"
)

type SeedFunc func(ctx context.Context, storage *db.Queries) error

var registry = map[string]SeedFunc{}

func Register(name string, fn SeedFunc) {
	registry[name] = fn
}

func Run(ctx context.Context, storage *db.Queries, name string) error {
	fn, ok := registry[name]
	if !ok {
		return fmt.Errorf("сид %q не найден. Доступные: %v", name, List())
	}
	slog.Info("запуск сида", slog.String("name", name))
	if err := fn(ctx, storage); err != nil {
		return fmt.Errorf("сид %q завершился с ошибкой: %w", name, err)
	}
	slog.Info("сид завершён", slog.String("name", name))
	return nil
}

func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
