package core

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"go.uber.org/fx"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/internal/plugin"
	servermeta "github.com/zalberix/cactus/apps/core/internal/server/meta"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
	"github.com/zalberix/cactus/apps/core/storage/db"
	lp "github.com/zalberix/cactus/libs/pipeline"
)

type Opts struct {
	fx.In
	Storage     Storage
	Broker      Broker
	FileStorage FileStorage
	Plugins     Plugins
	Meta        *servermeta.ServerMeta
}

func NewFx(opts Opts) *Service {
	return New(opts.Storage, opts.Broker, opts.FileStorage, opts.Plugins, opts.Meta)
}

//go:generate mockgen -package=mocks -destination=mocks/mock_storage.go github.com/zalberix/cactus/apps/core/internal/service/core Storage
type Storage interface {
	GetSystemById(ctx context.Context, id int32) (db.System, error)
	GetPriorityBySystemId(ctx context.Context, id int32) (db.GetPriorityBySystemIdRow, error)
	GetPriorityBySlug(ctx context.Context, slug string) (db.Priority, error)
	CreateManifest(ctx context.Context, value json.RawMessage) (db.Manifest, error)
	CreateMessage(ctx context.Context, arg db.CreateMessageParams) (db.Message, error)
	GetFilePathByUUID(ctx context.Context, argUUID uuid.UUID) (string, error)
	GetTypeWorkerBySlug(ctx context.Context, slug string) (db.TypeWorker, error)
	GetMessagesBy(ctx context.Context, arg db.GetMessagesByParams) ([]db.Message, error)
	GetStatusMessageByUUID(ctx context.Context, argUUID uuid.UUID) (string, error)
	GetTokenByPublicToken(ctx context.Context, publicToken string) (db.Token, error)
	GetTypeWorkers(ctx context.Context) ([]db.TypeWorker, error)
	GetKindWorkerBySlug(ctx context.Context, slug string) (db.KindWorker, error)
	CreateKindWorker(ctx context.Context, arg db.CreateKindWorkerParams) (db.KindWorker, error)
	CreateTypeWorker(ctx context.Context, arg db.CreateTypeWorkerParams) (db.TypeWorker, error)
	GetWorkerByUUID(ctx context.Context, argUUID uuid.UUID) (db.Worker, error)
	CreateWorker(ctx context.Context, arg db.CreateWorkerParams) (db.Worker, error)
	CreateFile(ctx context.Context, arg db.CreateFileParams) (db.File, error)
	CreatePipeline(ctx context.Context, arg db.CreatePipelineParams) (db.Pipeline, error)
	CreatePipelineStep(ctx context.Context, arg db.CreatePipelineStepParams) (db.PipelineStep, error)
	GetKindWokerByID(ctx context.Context, id int32) (db.KindWorker, error)
	GetTypeSlugWorkerByKindSlugWorker(ctx context.Context, slug string) (string, error)

	SetContext(ctx context.Context, db interface{}) error
	Rollback() error
	Commit() error
}

//go:generate mockgen -package=mocks -destination=mocks/mock_broker.go github.com/zalberix/cactus/apps/core/internal/service/core Broker
type Broker interface {
	EnsureStreamGroup(ctx context.Context, streamName, groupName string) error
	AddMessageToQueue(
		ctx context.Context,
		subject string,
		messageDTO lp.MessageInQueue,
		systemDTO lp.SystemInQueue,
		pipelineDTO lp.PipelineInQueue,
	) error
}

//go:generate mockgen -package=mocks -destination=mocks/mock_file_storage.go github.com/zalberix/cactus/apps/core/internal/service/core FileStorage
type FileStorage interface {
	Save(ctx context.Context, file multipart.File, ext mimetype.MIME) (path string, err error)
	Get(ctx context.Context, path string) (*os.File, error)
}

//go:generate mockgen -package=mocks -destination=mocks/mock_plugins.go github.com/zalberix/cactus/apps/core/internal/service/core Plugins
type Plugins interface {
	Add(slug string, plugin plugin.Plugin)
	Delete(slug string)
	Get(slug string) (p plugin.Plugin, ok bool)
}

//go:generate mockgen -package=mocks -destination=mocks/mock_pipeline_service.go github.com/zalberix/cactus/apps/core/internal/service/core PipelineService
type PipelineService interface {
	CreatePipelineTX(
		ctx context.Context,
		storageTx pipeline.StorageTx,
		arg pipeline.CreatePipelineParams,
	) ([]dto.Pipeline, error)
}

type Service struct {
	broker      Broker
	storage     Storage
	fileStorage FileStorage
	plugins     Plugins
	meta        *servermeta.ServerMeta
}

func New(
	storage Storage,
	broker Broker,
	fileStorage FileStorage,
	plugins Plugins,
	meta *servermeta.ServerMeta,
) *Service {
	return &Service{
		broker:      broker,
		storage:     storage,
		fileStorage: fileStorage,
		plugins:     plugins,
		meta:        meta,
	}
}
