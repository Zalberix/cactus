package seeds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SeedUsers вставляет demo-пользователя в тестовую организацию.
// Пароль: "password" (bcrypt hash).
func SeedUsers(ctx context.Context, db *sqlx.DB) error {
	// $2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi = "password"
	_, err := db.ExecContext(ctx, `
		INSERT INTO "user" (email, password, last_name, first_name, reset_password_after_login, organization_id)
		VALUES (
			'demo@test.local',
			'$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
			'Demo',
			'User',
			false,
			(SELECT id FROM organization WHERE code = 'test')
		)
		ON CONFLICT (email) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed demo user: %w", err)
	}
	return nil
}
