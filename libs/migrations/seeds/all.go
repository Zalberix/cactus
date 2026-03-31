package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedAll выполняет все сиды по порядку.
func SeedAll(ctx context.Context, db *pgxpool.Pool) error {
	ordered := []struct {
		name string
		fn   SQLSeedFunc
	}{
		{"organization", SeedOrganization},
		{"admin_user", SeedAdminUser},
		{"users", SeedUsers},
		{"permissions", SeedPermissions},
		{"admin_role", SeedAdminRole},
		{"work_types", SeedWorkTypes},
		{"systems", SeedSystems},
		{"demo", SeedDemo},
	}

	for _, s := range ordered {
		if err := s.fn(ctx, db); err != nil {
			return fmt.Errorf("сид %q: %w", s.name, err)
		}
	}

	return nil
}
