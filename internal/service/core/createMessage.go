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
	Plugin       plugin.Plugin
	IDKindWorker int32
	IDSystem     int32
	PrioritySlug string
	ChanelSlug   string
	Schema       any
	SendLater    *time.Time
	Files        []SetFileParams
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

	kindWorker, err := storage.GetKindWokerById(ctx, arg.IDKindWorker)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка получения вида воркера: %w", err)
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
			IDWorker: sql.NullInt32{
				Valid: false, // TODO: Пока NULL но надо определять по sendLater текущий Worker
			},
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
	piplines := []pl.Step{
		// pl.StepWaitSendQueue,  // TODO включать при send_later != nil
		pl.StepWaitQueue,
		pl.StepWork,
		pl.StepDone,
	}

	err = arg.Plugin.ExtendPipline(&piplines)
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка при расширении pipepline: %w", err)
	}
	pipeline, err := piplineService.CreatePipelineTX(ctx, tx, pipeline.CreatePipelineParams{
		Pipeline: piplines,
		Message:  dtoMessage.Message,
	})
	if err != nil {
		return dto.CreateMessage{}, fmt.Errorf("ошибка при создании pipepline: %w", err)
	}

	_ = pipeline

	s.AddMessageToQueue(ctx, AddMessageToQueueParams{
		SlugKindWorker:        kindWorker.Slug,
		Message:               newMessage,
		Files:                 files,
		SlugTypeWorker:        arg.ChanelSlug,
		WeightPriorityMessage: systemPriority.Weight + messagePriority.Weight,
	})

	err = tx.Commit()
	return dtoMessage, err
}
