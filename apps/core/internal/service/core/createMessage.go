package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/internal/plugin"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
	"github.com/zalberix/cactus/apps/core/storage/db"
	pl "github.com/zalberix/cactus/libs/pipeline"
)

type CreateMessageParams struct {
	Plugin         plugin.Plugin
	KindWorkerSlug string
	TypeWorkerSlug string
	IDSystem       int32
	PrioritySlug   string
	Schema         any
	SendLater      *time.Time
	Files          []SetFileParams
}

func (s *Service) CreateMessage(
	ctx context.Context,
	arg CreateMessageParams,
	piplineService PipelineService,
) (dto.CreateMessage, error) {
	storageTx := s.storage
	err := s.storage.SetContext(ctx, &storageTx)
	if err != nil {
		slog.Error("Ошибка создания контекста:", slog.String("error", err.Error()))
		return dto.CreateMessage{}, fmt.Errorf("невозможно создать транзакцию для сообщения")
	}
	defer storageTx.Rollback()

	var sendLater sql.NullTime
	if arg.SendLater == nil {
		sendLater = sql.NullTime{
			Valid: false,
		}
	} else {
		sendLater = sql.NullTime{
			Time:  *arg.SendLater,
			Valid: true,
		}
	}

	// system, err := storage.GetSystemById(ctx, arg.IDSystem)
	systemPriority, err := storageTx.GetPriorityBySystemId(ctx, arg.IDSystem)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения приоритета системы: %w", err)
	}

	messagePriority, err := storageTx.GetPriorityBySlug(ctx, arg.PrioritySlug)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения приоритета по slug: %w", err)
	}

	TypeWorker, err := storageTx.GetTypeWorkerBySlug(ctx, arg.TypeWorkerSlug)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения типа воркера по slug: %w", err)
	}

	value, err := json.Marshal(arg.Schema)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка создания json из данных: %w", err)
	}

	manifest, err := storageTx.CreateManifest(ctx, json.RawMessage(value))
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка создания манифеста: %w", err)
	}

	newMessage, err := storageTx.CreateMessage(
		ctx,
		db.CreateMessageParams{
			SystemID:   arg.IDSystem,
			ManifestID: manifest.ID,
			Uuid:       uuid.New(),
			Value:      value,
			Priority:   systemPriority.Weight + messagePriority.Weight,
			SendAt:     sendLater,
		},
	)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка создания сообщения: %w", err)
	}

	var files []dto.SetFile
	if len(arg.Files) != 0 {
		// TODO запараллелить сохранение файлов
		for _, file := range arg.Files {
			file.IDMessage = newMessage.ID
			ff, err := s.SetFileTX(ctx, storageTx, file)
			if err != nil {
				return dto.CreateMessage{}, fmt.Errorf("ошибка сохранения файла: %w", err)
			}
			files = append(files, ff)
		}
	}
	dtoMessage := dto.CreateMessage{
		Message: dto.Message{
			Message: newMessage,
			Value:   arg.Schema,
		},
		Files: &files,
	}

	stepInit, err := arg.Plugin.ExtendPipeline([]pl.Step{})
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка при расширении pipepline: %w", err)
	}
	pipelines, err := piplineService.CreatePipelineTX(ctx, storageTx, pipeline.CreatePipelineParams{
		Pipeline:  stepInit,
		Message:   dtoMessage.Message,
		ChannelID: TypeWorker.ID,
	})
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка при создании pipepline: %w", err)
	}

	slog.Info("AddMessageToQueueParams", slog.Any("AddMessageToQueueParams", AddMessageToQueueParams{
		SlugKindWorker:        arg.KindWorkerSlug,
		Message:               newMessage,
		Files:                 files,
		SlugTypeWorker:        arg.TypeWorkerSlug,
		WeightPriorityMessage: systemPriority.Weight + messagePriority.Weight,
		Step:                  pipelines[0].Step,
	}))

	s.AddMessageToQueue(ctx, AddMessageToQueueParams{
		SlugKindWorker:        arg.KindWorkerSlug,
		Message:               newMessage,
		Files:                 files,
		SlugTypeWorker:        arg.TypeWorkerSlug,
		WeightPriorityMessage: systemPriority.Weight + messagePriority.Weight,
		Step:                  pipelines[0].Step,
	})

	err = storageTx.Commit()
	return dtoMessage, err
}
