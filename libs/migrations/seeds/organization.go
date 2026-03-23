package seeds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SeedOrganization вставляет тестовую организацию.
func SeedOrganization(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO organization (name, code)
		VALUES ('Test Organization', 'test')
		ON CONFLICT (code) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed organization: %w", err)
	}
	return nil
}
