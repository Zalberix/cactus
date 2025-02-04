package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	dto "cactus/internal/DTO"
	"cactus/internal/storage/db"
)

type AddMessageToQueueParams struct {
	db.Message
	Files                 []dto.SetFile
	SlugKindWorker        string
	SlugTypeWorker        string
	WeightPriorityMessage int32
	Step                  int32
}

func (s *Service) AddMessageToQueue(ctx context.Context, arg AddMessageToQueueParams) error {
	nameQueue := fmt.Sprintf("messages:%v:%v:w-%v", arg.SlugTypeWorker, arg.SlugKindWorker, arg.WeightPriorityMessage)

	files := make([]dto.FileInMessageValueInMessageQueue, 0, len(arg.Files))
	for _, file := range arg.Files {
		f := dto.FileInMessageValueInMessageQueue{
			URL:  fmt.Sprintf("http://%s:%s/api/file/get?uuid=%s", s.meta.HostName, s.meta.Port, file.UUID.String()),
			Name: fmt.Sprintf("%v.%v", file.Title, file.Ext),
		}
		files = append(files, f)
	}

	dtoMessage := dto.MessageValueInMessageQueue{
		ID:       arg.ID,
		UUID:     arg.Uuid,
		Value:    arg.Value,
		CreateAt: arg.CreateAt,
		Files:    files,
	}

	if arg.SendLater.Valid {
		dtoMessage.SendLater = &arg.SendLater.Time
	}

	messageJSON, err := json.Marshal(dtoMessage)
	if err != nil {
		slog.Error("Не удалось сформировать JSON из сообщения: ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON из сообщения: %w", err)
	}

	keyType, err := s.rdb.Type(ctx, nameQueue).Result()
	if err != nil {
		slog.Error("ну удалось получить тип записи по ключу "+nameQueue+": ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON из сообщения: %w", err)
	}
	switch keyType {
	case "none":
		err = s.rdb.XGroupCreateMkStream(ctx, nameQueue, "reader", "0").Err()
		if err != nil {
			slog.Error("Не удалось создать группу для сообщений: ", slog.Any("err", err))
			return fmt.Errorf("не удалось создать группу для сообщений: %w", err)
		}
	case "stream":
		info, err := s.rdb.XInfoStream(ctx, nameQueue).Result()
		if err != nil {
			slog.Error("не удалось посмотреть информацию стриме: ", slog.Any("err", err))
			return fmt.Errorf("не удалось посмотреть информацию стриме: %w", err)
		}
		if info.Groups == 0 {
			err = s.rdb.XGroupCreate(ctx, nameQueue, "reader", "0").Err()
			if err != nil {
				slog.Error("Не удалось создать группу для сообщений: ", slog.Any("err", err))
				return fmt.Errorf("не удалось создать группу для сообщений: %w", err)
			}
		}
	default:
	}

	system, err := s.storage.GetSystemById(ctx, arg.IDSystem)
	if err != nil {
		slog.Error("невозможно получить систему по ID: ", slog.Any("err", err))
		return fmt.Errorf("невозможно получить систему по ID: %w", err)
	}

	systemJSON, err := json.Marshal(dto.SystemValueInMessageQueue{
		Name: system.Name,
	})
	if err != nil {
		slog.Error("Не удалось сформировать JSON для системы: ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON для системы: %w", err)
	}

	pipelineJSON, err := json.Marshal(dto.PipelineValueInMessageQueue{
		Step: arg.Step,
	})
	if err != nil {
		slog.Error("Не удалось сформировать JSON для системы: ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON для системы: %w", err)
	}

	err = s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: nameQueue,
		Values: map[string]interface{}{
			"message":  messageJSON,
			"system":   systemJSON,
			"pipeline": pipelineJSON,
		},
	}).Err()
	if err != nil {
		slog.Error("Не удалось добавить сообщение в стрим redis: ", slog.Any("err", err))
		return fmt.Errorf("не удалось добавить сообщение в стрим redis: %w", err)
	}

	return nil
}
