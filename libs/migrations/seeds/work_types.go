package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	smtpOrgBootstrapTokenHash     = "7fb1131c8efddfa50392b96c971ee3358d18a10c192eaf97e409b916558a22bc" //nolint:gosec // Seed stores SHA-256 hashes, not plaintext tokens.
	templateOrgBootstrapTokenHash = "16f657f9a9a798fe31dc35dd9b601499d56a285accfde0748e1af9a035f7d5fc" //nolint:gosec // Seed stores SHA-256 hashes, not plaintext tokens.
	telegramOrgBootstrapTokenHash = "79b8a204126030209d98fa2f53e730e3da31f891e7691e82e7c419d337b69d67" //nolint:gosec // Seed stores SHA-256 hashes, not plaintext tokens.
)

// SeedWorkTypes вставляет базовые типы работ (smtp, telegram, html).
// и создаёт dev bootstrap-токены для них.
func SeedWorkTypes(ctx context.Context, db *pgxpool.Pool) error {
	workTypes := []struct {
		Name string
		Code string
		Meta string
	}{
		{
			"Email SMTP", "smtp",
			`{"icon":"mail","color":"#3b82f6","category":"Channels"}`,
		},
		{
			"Telegram", "telegram",
			`{"icon":"message-square","color":"#0088cc","category":"Channels"}`,
		},
		{
			"HTML Template", "html",
			`{"icon":"file-code","color":"#10b981","category":"Content"}`,
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
	}

	orgTokens := []struct {
		WorkTypeCode     string
		Name             string
		TokenHash        string
		MaxActiveWorkers int
	}{
		{"smtp", "dev-smtp-org-1", smtpOrgBootstrapTokenHash, 100},
		{"html", "dev-template-org-1", templateOrgBootstrapTokenHash, 100},
		{"telegram", "dev-telegram-org-1", telegramOrgBootstrapTokenHash, 100},
	}

	for _, token := range orgTokens {
		_, err := db.Exec(ctx, `
			INSERT INTO worker_bootstrap_token (
				organization_id, work_type_id, name, token_hash, max_active_workers
			)
			SELECT org.id, wt.id, $3, $4, $5
			FROM organization org
			JOIN work_type wt ON wt.code = $2
			WHERE org.code = $1
			ON CONFLICT (token_hash) DO NOTHING
		`, defaultOrganizationCode, token.WorkTypeCode, token.Name, token.TokenHash, token.MaxActiveWorkers)
		if err != nil {
			return fmt.Errorf("seed org worker bootstrap token %s: %w", token.WorkTypeCode, err)
		}
	}

	return nil
}
