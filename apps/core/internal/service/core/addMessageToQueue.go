package core

import (
	"context"
	"fmt"
	"log/slog"

	dto "cactus/apps/core/internal/DTO"
	"cactus/apps/core/storage/db"
	"cactus/libs/shared/contracts"
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

	files := make([]contracts.FileInMessageValueInMessageQueue, 0, len(arg.Files))
	for _, file := range arg.Files {
		files = append(files, contracts.FileInMessageValueInMessageQueue{
			URL:  fmt.Sprintf("http://%s:%s/api/file/get?uuid=%s", s.meta.HostName, s.meta.Port, file.UUID.String()),
			Name: fmt.Sprintf("%v.%v", file.Title, file.Ext),
		})
	}

	dtoMessage := contracts.MessageValueInMessageQueue{
		ID:       arg.ID,
		UUID:     arg.Uuid,
		Value:    arg.Value,
		CreateAt: arg.CreatedAt,
		Files:    files,
	}

	if arg.SendAt.Valid {
		dtoMessage.SendLater = &arg.SendAt.Time
	}

	err := s.broker.EnsureStreamGroup(ctx, nameQueue, "reader")
	if err != nil {
		slog.Error("Ошибка при создании стрима и группы в Redis: ", slog.Any("err", err))
		return err
	}

	system, err := s.storage.GetSystemById(ctx, arg.SystemID)
	if err != nil {
		slog.Error("невозможно получить систему по ID: ", slog.Any("err", err))
		return fmt.Errorf("невозможно получить систему по ID: %w", err)
	}

	systemDTO := contracts.SystemValueInMessageQueue{Name: system.Name}
	pipelineDTO := contracts.PipelineValueInMessageQueue{Step: arg.Step}

	err = s.broker.AddMessageToQueue(ctx, nameQueue, dtoMessage, systemDTO, pipelineDTO)
	if err != nil {
		slog.Error("Ошибка при добавлении сообщения в очередь Redis: ", slog.Any("err", err))
		return err
	}

	return nil
}
