package core

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

func (s *Service) GetMessages(ctx context.Context, slug string, systemID int) ([]dto.Message, error) {
	typeWorker, err := s.storage.GetTypeWorkerBySlug(ctx, slug)
	if err != nil {
		return []dto.Message{}, err
	}

	messagesDB, err := s.storage.GetMessagesBy(ctx, db.GetMessagesByParams{
		IDTypeWorker: typeWorker.ID,
		IDSystem:     int32(systemID),
	})
	if err != nil {
		return []dto.Message{}, err
	}

	messages := make([]dto.Message, 0, len(messagesDB))

	for _, message := range messagesDB {
		// TODO тут плагин возвращается а не schema
		schema, exist := s.plugins.Get(slug)

		if !exist {
			return []dto.Message{}, fmt.Errorf("схема удалена: %v", slug)
		}

		if err := json.Unmarshal(message.Value, &schema); err != nil {
			return []dto.Message{}, fmt.Errorf("ошибка при разборе данных сообщения: %s", err.Error())
		}

		reflect.New(reflect.TypeOf(schema)).Elem()
		messages = append(messages, dto.Message{
			Message: message,
			Value:   schema,
		})
	}

	return messages, nil
}
