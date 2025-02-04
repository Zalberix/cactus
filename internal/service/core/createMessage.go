package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	dto "cactus/internal/DTO"
	pl "cactus/internal/pkg/pipeline"
	"cactus/internal/plugin"
	"cactus/internal/service/pipeline"
	"cactus/internal/storage/db"
)

type CreateMessageParams struct {
	Plugin         plugin.Plugin
	KindWorkerSlug string
	IDSystem       int32
	PrioritySlug   string
	ChanelSlug     string
	Schema         any
	SendLater      *time.Time
	Files          []SetFileParams
}

func (s *Service) CreateMessage(
	ctx context.Context,
	arg CreateMessageParams,
	piplineService pipeline.Service, // TODO переписать на interface
) (dto.CreateMessage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Ошибка создания контекста:", slog.String("error", err.Error()))
		return dto.CreateMessage{}, fmt.Errorf("невозможно создать транзакцию для сообщения")
	}
	defer tx.Rollback()

	storage := s.storage.WithTx(tx)

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
	systemPriority, err := storage.GetPriorityBySystemId(ctx, arg.IDSystem)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения приоритета системы: %w", err)
	}

	messagePriority, err := storage.GetPriorityBySlug(ctx, arg.PrioritySlug)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения приоритета по slug: %w", err)
	}

	TypeWorker, err := storage.GetTypeWorkerBySlug(ctx, arg.ChanelSlug)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения типа воркера по slug: %w", err)
	}

	value, err := json.Marshal(arg.Schema)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка создания json из данных: %w", err)
	}

	newMessage, err := storage.CreateMessage(
		ctx,
		db.CreateMessageParams{
			IDTypeWorker: TypeWorker.ID,
			IDSystem:     arg.IDSystem,
			Uuid:         uuid.New(),
			Value:        value,
			IDPriority:   messagePriority.ID,
			SendLater:    sendLater,
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
			ff, err := s.SetFileTX(ctx, tx, file)
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
	pipelines, err := piplineService.CreatePipelineTX(ctx, tx, pipeline.CreatePipelineParams{
		Pipeline: stepInit,
		Message:  dtoMessage.Message,
	})
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка при создании pipepline: %w", err)
	}

	s.AddMessageToQueue(ctx, AddMessageToQueueParams{
		SlugKindWorker:        arg.KindWorkerSlug,
		Message:               newMessage,
		Files:                 files,
		SlugTypeWorker:        arg.ChanelSlug,
		WeightPriorityMessage: systemPriority.Weight + messagePriority.Weight,
		Step:                  pipelines[0].Step,
	})

	err = tx.Commit()
	return dtoMessage, err
}
