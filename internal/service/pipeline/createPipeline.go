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

func (s *Service) CreatePipelineTX(ctx context.Context, tx *sql.Tx, arg CreatePipelineParams) ([]dto.Pipeline, error) {
	storage := s.storage
	if tx != nil {
		storage = storage.WithTx(tx)
	}

	createPipeline := make([]dto.Pipeline, 0, len(arg.Pipeline))
	if len(arg.Pipeline) == 0 {
		return []dto.Pipeline{}, fmt.Errorf("количество шагов не может быть меньше 1")
	}
	for _, step := range arg.Pipeline {
		createStep, err := storage.CreatePipelineStep(ctx, db.CreatePipelineStepParams{
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

// func (s *Service) GoToStepPipeline(
// 	ctx context.Context,
// 	uuidMeassage string,
// 	status pipeline.Status,
// 	step pipeline.Step,
// ) ([]dto.Pipeline, error) {
// 	uid, err := uuid.Parse(uuidMeassage)
// 	if err != nil {
// 		return []dto.Pipeline{}, err
// 	}

// 	pipelines, err := s.storage.GetAllPipelineByMessageUUID(ctx, uid)
// 	if err != nil {
// 		return []dto.Pipeline{}, err
// 	}

// 	tx, err := s.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return []dto.Pipeline{}, err
// 	}
// 	defer tx.Rollback()

// 	storage := s.storage.WithTx(tx)

// 	now := time.Now()
// 	var newPipelines []dto.Pipeline
// 	for _, p := range pipelines {
// 		var newPipeline db.Pipeline = p

// 		arg := db.UpdateStatusAndTimePipelineParams{
// 			ID:     p.ID,
// 			Status: string(status),
// 		}

// 		if p.Step < step.Step {
// 			if !p.TimeStart.Valid {
// 				arg.TimeStart = sql.NullTime{
// 					Valid: true,
// 					Time:  now,
// 				}
// 			}
// 			if !p.TimeEnd.Valid {
// 				arg.TimeEnd = sql.NullTime{
// 					Valid: true,
// 					Time:  now,
// 				}
// 			}
// 			newPipeline, err = storage.UpdateStatusAndTimePipeline(ctx, arg)
// 			if err != nil {
// 				return []dto.Pipeline{}, err
// 			}
// 			newPipelines = append(newPipelines, dto.Pipeline(newPipeline))
// 		} else if p.Step == step.Step {
// 			newPipeline, err = storage.UpdateStatusAndTimePipeline(ctx, db.UpdateStatusAndTimePipelineParams{
// 				ID:     p.ID,
// 				Status: string(status),
// 				TimeStart: sql.NullTime{
// 					Valid: true,
// 					Time:  now,
// 				},
// 				TimeEnd: sql.NullTime{
// 					Valid: false,
// 				},
// 			})
// 			if err != nil {
// 				return []dto.Pipeline{}, err
// 			}
// 		}
// 		newPipelines = append(newPipelines, dto.Pipeline(newPipeline))
// 	}

// 	if err := tx.Commit(); err != nil {
// 		return []dto.Pipeline{}, err
// 	}

// 	return newPipelines, nil
// }

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
