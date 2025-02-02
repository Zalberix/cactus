package pipeline

import (
	"context"
	"database/sql"
	"fmt"

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
	for i, step := range arg.Pipeline {
		createStep, err := storage.CreatePipelineStep(ctx, db.CreatePipelineStepParams{
			IDMessage: arg.Message.ID,
			Status:    string(pipeline.Wait),
			Step:      int32(i),
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

func (s *Service) StopPipeline(ctx context.Context) {
	_ = ctx
}

func (s *Service) NextStepPipeline(ctx context.Context) {
	_ = ctx
}

func (s *Service) BreakPipeline(ctx context.Context) {
	_ = ctx
}

func (s *Service) ErrorPipeline(ctx context.Context) {
	_ = ctx
}
