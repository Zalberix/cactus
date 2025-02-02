package core

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"cactus/internal/server/meta"
	"cactus/internal/storage/db"
	"cactus/internal/storage/file"
	"cactus/internal/storage/plugin"
)

type Service struct {
	db          *sqlx.DB
	rdb         *redis.Client
	storage     *db.Queries
	fileStorage *file.Storage
	plugins     *plugin.Storage
	meta        *meta.ServerMeta
}

func New(
	db *sqlx.DB,
	storage *db.Queries,
	rdb *redis.Client,
	fileStorage *file.Storage,
	plugins *plugin.Storage,
	meta *meta.ServerMeta,
) *Service {
	return &Service{
		db:          db,
		rdb:         rdb,
		storage:     storage,
		fileStorage: fileStorage,
		plugins:     plugins,
		meta:        meta,
	}
}
