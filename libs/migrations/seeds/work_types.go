package seeds

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zalberix/cactus/apps/core/storage/db"
)

func init() {
	Register("work_types", SeedWorkTypes)
}

func SeedWorkTypes(ctx context.Context, storage *db.Queries) error {
	workTypes := []db.CreateWorkTypeParams{
		{Name: "Email рассылка", Code: "email", Description: sql.NullString{Valid: true, String: "Отправка email сообщений через SMTP/API провайдеров"}},
		{Name: "Telegram уведомления", Code: "telegram", Description: sql.NullString{Valid: true, String: "Отправка сообщений через Telegram Bot API"}},
		{Name: "SMS сообщения", Code: "sms", Description: sql.NullString{Valid: true, String: "Отправка SMS через провайдеров"}},
		{Name: "Push уведомления", Code: "push", Description: sql.NullString{Valid: true, String: "Push уведомления на мобильные устройства"}},
		{Name: "In-app уведомления", Code: "in_app", Description: sql.NullString{Valid: true, String: "Доставка уведомлений через WebSocket gateway"}},
	}

	for _, wt := range workTypes {
		_, err := storage.CreateWorkType(ctx, wt)
		if err != nil {
			return fmt.Errorf("ошибка при создании типа работ %q: %w", wt.Code, err)
		}
	}

	return nil
}
