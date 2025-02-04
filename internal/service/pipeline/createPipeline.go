package pipeline

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	dto "cactus/internal/DTO"
	"cactus/internal/pkg/pipeline"
	"cactus/internal/storage/db"
)

type CreatePipelineParams struct {
	Pipeline []pipeline.Step
	Message  dto.Message
}

func (s *Service) CreatePipeline(ctx context.Context, arg CreatePipelineParams) ([]dto.Pipeline, error) {
	return s.CreatePipelineTX(ctx, nil, arg)
}

func (s *Service) CreatePipelineTX(ctx context.Context, storageTx interface {
	CreatePipelineStep(ctx context.Context, arg db.CreatePipelineStepParams) (db.Pipeline, error)
}, arg CreatePipelineParams,
) ([]dto.Pipeline, error) {
	createPipeline := make([]dto.Pipeline, 0, len(arg.Pipeline))
	if len(arg.Pipeline) == 0 {
		return []dto.Pipeline{}, fmt.Errorf("количество шагов не может быть меньше 1")
	}
	for _, step := range arg.Pipeline {
		createStep, err := storageTx.CreatePipelineStep(ctx, db.CreatePipelineStepParams{
			IDMessage: arg.Message.ID,
			Status:    string(pipeline.Wait),
			Step:      step.Step,
			Name:      step.Name,
			TimeStart: sql.NullTime{
				Valid: false,
			},
			TimeEnd: sql.NullTime{
				Valid: false,
			},
		})
		if err != nil {
			return []dto.Pipeline{}, fmt.Errorf("ошибка создания pipeline шага для %v", step.Name)
		}
		createPipeline = append(createPipeline, dto.Pipeline(createStep))
	}

	// if !arg.Message.SendLater.Valid {

	// }

	return createPipeline, nil
}

func (s *Service) UpdateStatusPipeline(
	ctx context.Context,
	uuidMessage string,
	step int32,
	status pipeline.Status,
	workerUUID string,
) (dto.Pipeline, error) {
	UUIDm, err := uuid.Parse(uuidMessage)
	if err != nil {
		return dto.Pipeline{}, fmt.Errorf("проверьте UUID сообщения, он неверного формата: %w", err)
	}

	UUIDw, err := uuid.Parse(workerUUID)
	if err != nil {
		return dto.Pipeline{}, fmt.Errorf("проверьте UUID воркера, он неверного формата: %w", err)
	}

	idPipeline, err := s.storage.GetIdPipelineByUUIDMessageAndStep(ctx, db.GetIdPipelineByUUIDMessageAndStepParams{
		Uuid: UUIDm,
		Step: step,
	})
	if err != nil {
		return dto.Pipeline{}, fmt.Errorf("шаг пайплайна не найден по uuid сообщения и номеру шага: %w", err)
	}

	worker, err := s.storage.GetWorkerByUUID(ctx, UUIDw)
	if err != nil {
		return dto.Pipeline{}, fmt.Errorf(
			"воркер с uuid %v не зарегистрирован не найден по uuid сообщения и номеру шага: %w",
			UUIDw.String(),
			err,
		)
	}

	newPipeline, err := s.storage.UpdatePipelineStatusAndWorkerByID(ctx, db.UpdatePipelineStatusAndWorkerByIDParams{
		ID:     idPipeline,
		Status: string(status),
		IDWorker: sql.NullInt32{
			Valid: true,
			Int32: worker.ID,
		},
	})
	if err != nil {
		return dto.Pipeline{}, fmt.Errorf("не удалось обновить pipeline: %w", err)
	}

	return dto.Pipeline(newPipeline), nil
}
