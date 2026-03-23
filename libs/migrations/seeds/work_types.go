package seeds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SeedWorkTypes вставляет базовые типы работ (smtp, telegram).
func SeedWorkTypes(ctx context.Context, db *sqlx.DB) error {
	workTypes := []struct {
		Name string
		Code string
	}{
		{"Email SMTP", "smtp"},
		{"Telegram", "telegram"},
	}

	for _, wt := range workTypes {
		_, err := db.ExecContext(ctx, `
			INSERT INTO work_type (name, code)
			VALUES ($1, $2)
			ON CONFLICT (code) DO NOTHING
		`, wt.Name, wt.Code)
		if err != nil {
			return fmt.Errorf("seed work type %s: %w", wt.Code, err)
		}
	}
	return nil
}
