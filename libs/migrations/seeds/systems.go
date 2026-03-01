package seeds

import (
	"context"
	"database/sql"
	"fmt"

	"cactus/apps/core/storage/db"
)

func init() {
	Register("systems", SeedSystems)
}

func SeedSystems(ctx context.Context, storage *db.Queries) error {
	user, err := storage.GetUserByEmail(ctx, "demo@mail.ru")
	if err != nil {
		return fmt.Errorf("не найден пользователь demo@mail.ru: %w", err)
	}

	channel, err := storage.GetChannelBySlug(ctx, "email")
	if err != nil {
		return fmt.Errorf("не найден канал email: %w", err)
	}

	system, err := storage.CreateSystem(ctx, db.CreateSystemParams{
		UserCreatorID: sql.NullInt32{Valid: true, Int32: user.ID},
		Name:          "Тестовая система",
		Description:   sql.NullString{Valid: true, String: "demo для работы с системой"},
		IsActive:      true,
		Priority:      0,
		PublicToken:    sql.NullString{Valid: true, String: "12345678910"},
		PrivateToken:  sql.NullString{Valid: true, String: "10987654321"},
	})
	if err != nil {
		return fmt.Errorf("ошибка при создании системы: %w", err)
	}

	err = storage.AddChannelForSystem(ctx, db.AddChannelForSystemParams{
		SystemID:  system.ID,
		ChannelID: channel.ID,
	})
	if err != nil {
		return fmt.Errorf("ошибка при привязке канала к системе: %w", err)
	}

	return nil
}
