package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedAdminRole создаёт роль "admin" со всеми правами и назначает её admin-пользователю.
func SeedAdminRole(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("seed admin role: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Получаем organization_id
	var orgID int
	err = tx.QueryRow(ctx, `SELECT id FROM organization WHERE code = 'test' LIMIT 1`).Scan(&orgID)
	if err != nil {
		return fmt.Errorf("seed admin role: get org: %w", err)
	}

	// Получаем user_id
	var userID int
	err = tx.QueryRow(ctx, `SELECT id FROM "user" WHERE email = 'admin@test.local' LIMIT 1`).Scan(&userID)
	if err != nil {
		return fmt.Errorf("seed admin role: get user: %w", err)
	}

	// Создаём или находим роль admin
	var roleID int
	err = tx.QueryRow(ctx, `
		SELECT id FROM role WHERE organization_id = $1 AND name = 'admin' LIMIT 1
	`, orgID).Scan(&roleID)
	if err != nil {
		// Роль не найдена — создаём
		err = tx.QueryRow(ctx, `
			INSERT INTO role (organization_id, name, description, is_system)
			VALUES ($1, 'admin', 'Full access administrator role', true)
			RETURNING id
		`, orgID).Scan(&roleID)
		if err != nil {
			return fmt.Errorf("seed admin role: create role: %w", err)
		}
	}

	// Привязываем все пермишены к роли
	_, err = tx.Exec(ctx, `
		INSERT INTO permission_role (permission_id, role_id)
		SELECT p.id, $1 FROM permission p
		ON CONFLICT DO NOTHING
	`, roleID)
	if err != nil {
		return fmt.Errorf("seed admin role: assign permissions: %w", err)
	}

	// Привязываем роль к юзеру
	_, err = tx.Exec(ctx, `
		INSERT INTO role_user (role_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, roleID, userID)
	if err != nil {
		return fmt.Errorf("seed admin role: assign role to user: %w", err)
	}

	return tx.Commit(ctx)
}
