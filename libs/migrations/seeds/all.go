package seeds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SeedAll выполняет все сиды по порядку.
func SeedAll(ctx context.Context, db *sqlx.DB) error {
	ordered := []struct {
		name string
		fn   SQLSeedFunc
	}{
		{"organization", SeedOrganization},
		{"admin_user", SeedAdminUser},
		{"users", SeedUsers},
		{"permissions", SeedPermissions},
		{"work_types", SeedWorkTypes},
		{"systems", SeedSystems},
	}

	for _, s := range ordered {
		if err := s.fn(ctx, db); err != nil {
			return fmt.Errorf("сид %q: %w", s.name, err)
		}
	}

	return nil
}
