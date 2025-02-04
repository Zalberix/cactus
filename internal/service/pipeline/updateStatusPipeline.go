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
