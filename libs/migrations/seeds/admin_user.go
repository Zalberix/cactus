package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedAdminUser вставляет admin-пользователя в тестовую организацию.
// Пароль: "admin" (bcrypt hash).
func SeedAdminUser(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		INSERT INTO "user" (email, password, last_name, first_name, reset_password_after_login, organization_id)
		VALUES (
			'admin@test.local',
			'$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
			'Admin',
			'System',
			false,
			(SELECT id FROM organization WHERE code = 'test')
		)
		ON CONFLICT (email) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}
	return nil
}
