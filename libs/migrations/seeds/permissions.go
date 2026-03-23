package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zalberix/cactus/libs/permissions"
)

// SeedPermissions вставляет базовые права доступа.
func SeedPermissions(ctx context.Context, db *pgxpool.Pool) error {
	for _, p := range permissions.All() {
		_, err := db.Exec(ctx, `
			INSERT INTO permission (slug, is_private)
			VALUES ($1, $2)
			ON CONFLICT (slug) DO NOTHING
		`, p.String(), false)
		if err != nil {
			return fmt.Errorf("seed permission %s: %w", p, err)
		}
	}
	return nil
}
