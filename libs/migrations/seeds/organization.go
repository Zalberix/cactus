package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedOrganization вставляет тестовую организацию.
func SeedOrganization(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		INSERT INTO organization (name, code)
		VALUES ('Test Organization', 'test')
		ON CONFLICT (code) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed organization: %w", err)
	}
	return nil
}
