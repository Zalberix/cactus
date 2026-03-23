package seeds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/zalberix/cactus/libs/permissions"
)

// SeedPermissions вставляет базовые права доступа.
func SeedPermissions(ctx context.Context, db *sqlx.DB) error {
	for _, p := range permissions.All() {
		_, err := db.ExecContext(ctx, `
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
