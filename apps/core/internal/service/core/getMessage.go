package core

import (
	"context"

	"github.com/zalberix/cactus/apps/core/storage/db"
)

func (s *Service) GetMessage(ctx context.Context) (db.Message, error) {
	_ = ctx
	return db.Message{}, nil
}
