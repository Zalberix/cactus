package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const htmlBootstrapTokenHash = "fcaf410de830d072b301a2e7ca1092320647584c2585c9b5e5f64cebf1990164"

// SeedWorkTypes вставляет базовые типы работ (smtp, telegram, control-типы)
// и создаёт dev bootstrap-токены для них.
func SeedWorkTypes(ctx context.Context, db *pgxpool.Pool) error {
	workTypes := []struct {
		Name      string
		Code      string
		Meta      string
		TokenHash string // sha256(plaintext)
	}{
		// plaintext: "dev-smtp-bootstrap-token"
		{
			"Email SMTP", "smtp",
			`{"icon":"mail","color":"#3b82f6","category":"Channels"}`,
			"224962dd1073f04c98ccd8f27fe21a9c1046619649fad79743bd41c2fc834118",
		},
		// plaintext: "dev-telegram-bootstrap-token"
		{
			"Telegram", "telegram",
			`{"icon":"message-square","color":"#0088cc","category":"Channels"}`,
			"",
		},
		// plaintext: "dev-html-bootstrap-token"
		{
			"HTML Template", "html",
			`{"icon":"file-code","color":"#10b981","category":"Content"}`,
			htmlBootstrapTokenHash,
		},
	}

	for _, wt := range workTypes {
		_, err := db.Exec(ctx, `
			INSERT INTO work_type (name, code, meta)
			VALUES ($1, $2, $3::jsonb)
			ON CONFLICT (code) DO UPDATE SET meta = $3::jsonb
		`, wt.Name, wt.Code, wt.Meta)
		if err != nil {
			return fmt.Errorf("seed work type %s: %w", wt.Code, err)
		}

		if wt.TokenHash == "" {
			continue
		}

		_, err = db.Exec(ctx, `
			INSERT INTO work_type_token (work_type_id, token_hash, is_active)
			SELECT wt.id, $2::text, true
			FROM work_type wt
			WHERE wt.code = $1
			ON CONFLICT (work_type_id) DO NOTHING
		`, wt.Code, wt.TokenHash)
		if err != nil {
			return fmt.Errorf("seed work type token %s: %w", wt.Code, err)
		}
	}
	return nil
}
