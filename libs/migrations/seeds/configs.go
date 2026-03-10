package seeds

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/pqtype"

	"github.com/zalberix/cactus/apps/core/storage/db"
	"github.com/zalberix/cactus/libs/shared/configschema"
)

func init() {
	Register("configs", SeedConfigs)
}

func SeedConfigs(ctx context.Context, storage *db.Queries) error {
	schema, err := json.Marshal([]configschema.ConfigField{
		{Type: "numeric", Slug: "host", Name: "ip адрес сервера SMTP"},
		{Type: "numeric", Slug: "port", Name: "Порт сервера"},
		{Type: "text", Slug: "from", Name: "Адрес отправителя"},
	})
	if err != nil {
		return fmt.Errorf("ошибка сериализации config_schema: %w", err)
	}

	cfg, err := json.Marshal(map[string]string{
		"host": "127.0.0.1",
		"port": "1025",
		"from": "cactus@gmail.com",
	})
	if err != nil {
		return fmt.Errorf("ошибка сериализации config: %w", err)
	}

	_, err = storage.CreateConfig(ctx, db.CreateConfigParams{
		Name:         "smtp сервер",
		ConfigSchema: schema,
		Config: pqtype.NullRawMessage{
			RawMessage: cfg,
			Valid:      true,
		},
	})
	if err != nil {
		return fmt.Errorf("ошибка при создании конфигурации: %w", err)
	}

	return nil
}
