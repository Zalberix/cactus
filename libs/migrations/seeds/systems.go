package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedSystems вставляет тестовую систему, привязанную к тестовой организации и admin-пользователю.
func SeedSystems(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		INSERT INTO system (organization_id, user_creator_id, name, description, is_active, priority)
		VALUES (
			(SELECT id FROM organization WHERE code = 'test'),
			(SELECT id FROM "user" WHERE email = 'admin@test.local'),
			'Тестовая система',
			'Demo-система для разработки',
			true,
			0
		)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed system: %w", err)
	}
	return nil
}
