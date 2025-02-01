package core

import (
	dto "cactus/internal/DTO"
	"cactus/internal/storage/db"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type AddMessageToQueueParams struct {
	db.Message
	Files                 []dto.SetFile
	SlugKindWorker        string
	SlugTypeWorker        string
	WeightPriorityMessage int32
}

func (s *Service) AddMessageToQueue(ctx context.Context, arg AddMessageToQueueParams) error {
	nameQueue := fmt.Sprintf("messages:%v:%v:w-%v", arg.SlugTypeWorker, arg.SlugKindWorker, arg.WeightPriorityMessage)

	var files []dto.FileInMessageValueInMessageQueue
	for _, file := range arg.Files {
		f := dto.FileInMessageValueInMessageQueue{
			Url:  fmt.Sprintf("http://%s:%s/api/file/get?uuid=%s", s.meta.HostName, s.meta.Port, file.UUID.String()),
			Name: fmt.Sprintf("%v.%v", file.Title, file.Ext),
		}
		files = append(files, f)
	}

	dtoMessage := dto.MessageValueInMessageQueue{
		ID:       arg.ID,
		Uuid:     arg.Uuid,
		Value:    arg.Value,
		CreateAt: arg.CreateAt,
		Files:    files,
	}

	if arg.SendLater.Valid {
		dtoMessage.SendLater = &arg.SendLater.Time
	}

	messageJson, err := json.Marshal(dtoMessage)
	if err != nil {
		slog.Error("Не удалось сформировать JSON из сообщения: ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON из сообщения: %w", err)
	}

	keyType, err := s.rdb.Type(ctx, nameQueue).Result()
	if err != nil {
		slog.Error("ну удалось получить тип записи по ключу "+nameQueue+": ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON из сообщения: %w", err)
	}
	if keyType == "stream" {
		info, err := s.rdb.XInfoStream(ctx, nameQueue).Result()
		if err != nil {

		}
		if info.Groups == 0 {
			err = s.rdb.XGroupCreate(ctx, nameQueue, "reader", "0").Err()
			if err != nil {
				slog.Error("Не удалось создать группу для сообщений: ", slog.Any("err", err))
				return fmt.Errorf("не удалось создать группу для сообщений: %w", err)
			}
		}
	} else if keyType == "none" {
		err = s.rdb.XGroupCreateMkStream(ctx, nameQueue, "reader", "0").Err()
		if err != nil {
			slog.Error("Не удалось создать группу для сообщений: ", slog.Any("err", err))
			return fmt.Errorf("не удалось создать группу для сообщений: %w", err)
		}
	}

	system, err := s.storage.GetSystemById(ctx, arg.IDSystem)
	if err != nil {
		slog.Error("невозможно получить систему по ID: ", slog.Any("err", err))
		return fmt.Errorf("невозможно получить систему по ID: %w", err)
	}

	systemJson, err := json.Marshal(dto.SystemValueInMessageQueue{
		Name: system.Name,
	})
	if err != nil {
		slog.Error("Не удалось сформировать JSON для системы: ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON для системы: %w", err)
	}

	err = s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: nameQueue,
		Values: map[string]interface{}{
			"message": messageJson,
			"system":  systemJson,
		},
	}).Err()
	if err != nil {
		slog.Error("Не удалось добавить сообщение в стрим redis: ", slog.Any("err", err))
		return fmt.Errorf("не удалось добавить сообщение в стрим redis: %w", err)
	}

	return nil
}
