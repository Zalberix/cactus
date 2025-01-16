package pipeline

import (
	"cactus/internal/storage/db"
	"cactus/internal/storage/plugin"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	db      *sqlx.DB
	rdb     *redis.Client
	storage *db.Queries
	plugins *plugin.Storage
}

func New(db *sqlx.DB, storage *db.Queries, rdb *redis.Client, plugins *plugin.Storage) *Service {
	return &Service{
		db:      db,
		rdb:     rdb,
		storage: storage,
		plugins: plugins,
	}
}
