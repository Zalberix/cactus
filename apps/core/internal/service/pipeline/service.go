package pipeline

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/internal/pkg/wshub"
	"github.com/zalberix/cactus/apps/core/internal/storage/plugin"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

type Opts struct {
	fx.In
	Storage Storage
	Plugins *plugin.Storage
}

func NewFx(opts Opts) *Service {
	return New(opts.Storage, opts.Plugins)
}

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
	GetIDPipelineByUUIDMessageAndStep(ctx context.Context, arg db.GetIDPipelineByUUIDMessageAndStepParams) (int32, error)
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
