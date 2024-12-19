package core

import (
	"cactus/internal/storage/db"
	filestorage "cactus/internal/storage/file"

	"github.com/jmoiron/sqlx"
)

type Storage interface {
}

type Service struct {
	db          *sqlx.DB
	storage     *db.Queries
	fileStorage *filestorage.FileStorage
}

func New(db *sqlx.DB, storage *db.Queries, fileStorage *filestorage.FileStorage) *Service {
	return &Service{
		db:          db,
		storage:     storage,
		fileStorage: fileStorage,
	}
}
