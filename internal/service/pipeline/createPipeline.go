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

func (s *Service) CreatePipelineTX(
	ctx context.Context,
	storageTx StorageTx,
	arg CreatePipelineParams,
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

	return createPipeline, nil
}
