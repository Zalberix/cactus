package seeds

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

// SQLSeedFunc — функция сида, работающая с сырым SQL через sqlx.DB.
type SQLSeedFunc func(ctx context.Context, db *sqlx.DB) error

var registry = map[string]SQLSeedFunc{}

func Register(name string, fn SQLSeedFunc) {
	registry[name] = fn
}

func Run(ctx context.Context, db *sqlx.DB, name string) error {
	fn, ok := registry[name]
	if !ok {
		return fmt.Errorf("сид %q не найден. Доступные: %v", name, List())
	}
	slog.Info("запуск сида", slog.String("name", name))
	if err := fn(ctx, db); err != nil {
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
