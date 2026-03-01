package core

import (
	"context"

	"cactus/apps/core/storage/db"
)

func (s *Service) GetTokenByPublicToken(ctx context.Context, token string) (db.Token, error) {
	dbToken, err := s.storage.GetTokenByPublicToken(ctx, token)
	return dbToken, err
}
