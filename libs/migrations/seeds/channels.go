package seeds

import (
	"context"
	"fmt"

	"cactus/apps/core/storage/db"
)

func init() {
	Register("channels", SeedChannels)
}

func SeedChannels(ctx context.Context, storage *db.Queries) error {
	channels := []db.CreateChannelParams{
		{Slug: "email", Name: "Рассылка писем"},
		{Slug: "telegram", Name: "Телеграмм уведомления"},
		{Slug: "push", Name: "Push уведомления"},
		{Slug: "sms", Name: "SMS-сообщения"},
		{Slug: "oneC", Name: "1С"},
	}

	for _, ch := range channels {
		_, err := storage.CreateChannel(ctx, ch)
		if err != nil {
			return fmt.Errorf("ошибка при создании канала %q: %w", ch.Slug, err)
		}
	}

	return nil
}
