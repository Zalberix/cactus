package pipeline

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"cactus/internal/pkg/wshub"
	"cactus/internal/storage/db"
	"cactus/internal/storage/plugin"
)

type Service struct {
	db      *sqlx.DB
	rdb     *redis.Client
	storage *db.Queries
	plugins *plugin.Storage
	wshub   *wshub.PipelineHub
}

func New(
	db *sqlx.DB,
	storage *db.Queries,
	rdb *redis.Client,
	plugins *plugin.Storage,
) *Service {
	return &Service{
		db:      db,
		rdb:     rdb,
		storage: storage,
		plugins: plugins,
	}
}

func (s *Service) RunWS(ctx context.Context) {
	pipelineHub := wshub.NewPipelineHub(ctx, s.rdb, s)
	s.wshub = pipelineHub
	go pipelineHub.Run()
}

func (s *Service) Hub() *wshub.PipelineHub {
	return s.wshub
}
