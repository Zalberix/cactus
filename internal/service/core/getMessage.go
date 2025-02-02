package core

import (
	"context"

	"cactus/internal/storage/db"
)

func (s *Service) GetMessage(ctx context.Context) (db.Message, error) {
	_ = ctx
	return db.Message{}, nil
}
