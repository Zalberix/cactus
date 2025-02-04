package pipeline

import (
	"context"

	"github.com/google/uuid"

	"cactus/internal/pkg/wshub"
	"cactus/internal/storage/db"
	"cactus/internal/storage/plugin"
)

type Service struct {
	storage Storage
	plugins *plugin.Storage
	wshub   *wshub.PipelineHub
}

type Storage interface {
	UpdatePipelineStatusAndWorkerByID(
		ctx context.Context,
		arg db.UpdatePipelineStatusAndWorkerByIDParams,
	) (db.Pipeline, error)
	GetWorkerByUUID(ctx context.Context, argUUID uuid.UUID) (db.Worker, error)
	GetIdPipelineByUUIDMessageAndStep(ctx context.Context, arg db.GetIdPipelineByUUIDMessageAndStepParams) (int32, error)
}

func New(
	storage Storage,
	plugins *plugin.Storage,
) *Service {
	return &Service{
		storage: storage,
		plugins: plugins,
	}
}

func (s *Service) SetHub(hub *wshub.PipelineHub) {
	s.wshub = hub
}

func (s *Service) Hub() *wshub.PipelineHub {
	return s.wshub
}
