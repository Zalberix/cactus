package pipeline

import (
	"context"
	"database/sql"
	"fmt"

	dto "cactus/apps/core/internal/DTO"
	"cactus/libs/shared/pipeline"
	"cactus/apps/core/storage/db"
)

type CreatePipelineParams struct {
	Pipeline  []pipeline.Step
	Message   dto.Message
	ChannelID int32
}

func (s *Service) CreatePipeline(ctx context.Context, arg CreatePipelineParams) ([]dto.Pipeline, error) {
	return s.CreatePipelineTX(ctx, nil, arg)
}

func (s *Service) CreatePipelineTX(
	ctx context.Context,
	storageTx StorageTx,
	arg CreatePipelineParams,
) ([]dto.Pipeline, error) {
	if len(arg.Pipeline) == 0 {
		return []dto.Pipeline{}, fmt.Errorf("количество шагов не может быть меньше 1")
	}

	parentPipeline, err := storageTx.CreatePipeline(ctx, db.CreatePipelineParams{
		MessageID:        arg.Message.ID,
		ParentPipelineID: sql.NullInt32{Valid: false},
	})
	if err != nil {
		return []dto.Pipeline{}, fmt.Errorf("ошибка создания pipeline: %w", err)
	}

	createPipeline := make([]dto.Pipeline, 0, len(arg.Pipeline))
	for _, step := range arg.Pipeline {
		createStep, err := storageTx.CreatePipelineStep(ctx, db.CreatePipelineStepParams{
			PipelineID:           parentPipeline.ID,
			WorkerID:             sql.NullInt32{Valid: false},
			ChannelID:            arg.ChannelID,
			Step:                 step.Step,
			TimeStart:            sql.NullTime{Valid: false},
			TimeEnd:              sql.NullTime{Valid: false},
			PipelineStepStatusID: 1, // wait
		})
		if err != nil {
			return []dto.Pipeline{}, fmt.Errorf("ошибка создания pipeline шага для %v", step.Name)
		}
		createPipeline = append(createPipeline, dto.Pipeline(createStep))
	}

	return createPipeline, nil
}
