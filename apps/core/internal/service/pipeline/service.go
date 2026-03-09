package pipeline

import (
	"context"

	"github.com/google/uuid"

	"github.com/zalberix/cactus/apps/core/internal/pkg/wshub"
	"github.com/zalberix/cactus/apps/core/storage/db"
	"github.com/zalberix/cactus/apps/core/internal/storage/plugin"
)

type Service struct {
	storage Storage
	plugins *plugin.Storage
	wshub   *wshub.PipelineHub
}

//go:generate mockgen -package=mocks -destination=mocks/mock_storage_tx.go cactus/internal/service/pipeline StorageTx
type StorageTx interface {
	CreatePipeline(ctx context.Context, arg db.CreatePipelineParams) (db.Pipeline, error)
	CreatePipelineStep(ctx context.Context, arg db.CreatePipelineStepParams) (db.PipelineStep, error)
}

//go:generate mockgen -package=mocks -destination=mocks/mock_storage.go cactus/internal/service/pipeline Storage
type Storage interface {
	UpdatePipelineStatusAndWorkerByID(
		ctx context.Context,
		arg db.UpdatePipelineStatusAndWorkerByIDParams,
	) (db.PipelineStep, error)
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
