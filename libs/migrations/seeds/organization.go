package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultOrganizationCode = "test"

// SeedOrganization вставляет тестовую организацию.
func SeedOrganization(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		INSERT INTO organization (name, code)
		VALUES ('Организация', $1)
		ON CONFLICT (code) DO NOTHING
	`, defaultOrganizationCode)
	if err != nil {
		return fmt.Errorf("seed organization: %w", err)
	}
	return nil
}
