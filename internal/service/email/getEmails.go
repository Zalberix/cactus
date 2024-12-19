package email

import (
	dto "cactus/internal/DTO"
	"cactus/internal/schema"
	"cactus/internal/storage/db"
	"context"
	"encoding/json"
	"fmt"
)

func (s *Service) GetEmails(ctx context.Context, systemID int) ([]dto.Message[schema.Email], error) {

	typeWorker, err := s.storage.GetTypeWorkerBySlug(ctx, SLUG_TYPE_WORKER)
	if err != nil {
		return []dto.Message[schema.Email]{}, err
	}

	messages, err := s.storage.GetMessagesBy(ctx, db.GetMessagesByParams{
		IDTypeWorker: typeWorker.ID,
		IDSystem:     int32(systemID),
	})

	if err != nil {
		return []dto.Message[schema.Email]{}, err
	}

	var emails []dto.Message[schema.Email]

	for _, message := range messages {

		var value schema.Email
		if err := json.Unmarshal(message.Value, &value); err != nil {
			return []dto.Message[schema.Email]{}, fmt.Errorf("ошибка при разборе данных сообщения: %s", err.Error())
		}

		emails = append(emails, dto.Message[schema.Email]{
			Message: message,
			Value:   value,
		})
	}

	return emails, nil
}
