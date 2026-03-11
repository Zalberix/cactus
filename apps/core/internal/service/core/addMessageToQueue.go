package core

import (
	"context"
	"fmt"
	"log/slog"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/storage/db"
	lp "github.com/zalberix/cactus/libs/pipeline"
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
	subject := fmt.Sprintf("messages.%v.%v.w-%v", arg.SlugTypeWorker, arg.SlugKindWorker, arg.WeightPriorityMessage)

	files := make([]lp.FileInMessage, 0, len(arg.Files))
	for _, file := range arg.Files {
		files = append(files, lp.FileInMessage{
			URL:  fmt.Sprintf("http://%s:%s/api/file/get?uuid=%s", s.meta.HostName, s.meta.Port, file.UUID.String()),
			Name: fmt.Sprintf("%v.%v", file.Title, file.Ext),
		})
	}

	dtoMessage := lp.MessageInQueue{
		ID:       arg.ID,
		UUID:     arg.Uuid,
		Value:    arg.Value,
		CreateAt: arg.CreatedAt,
		Files:    files,
	}

	if arg.SendAt.Valid {
		dtoMessage.SendLater = &arg.SendAt.Time
	}

	if err := s.broker.EnsureStreamGroup(ctx, subject, "reader"); err != nil {
		slog.Error("Ошибка при создании стрима NATS:", slog.Any("err", err))
		return err
	}

	system, err := s.storage.GetSystemByID(ctx, arg.SystemID)
	if err != nil {
		slog.Error("невозможно получить систему по ID:", slog.Any("err", err))
		return fmt.Errorf("невозможно получить систему по ID: %w", err)
	}

	systemDTO := lp.SystemInQueue{Name: system.Name}
	pipelineDTO := lp.PipelineInQueue{Step: arg.Step}

	if err = s.broker.AddMessageToQueue(ctx, subject, dtoMessage, systemDTO, pipelineDTO); err != nil {
		slog.Error("Ошибка при добавлении сообщения в очередь NATS:", slog.Any("err", err))
		return err
	}

	return nil
}
