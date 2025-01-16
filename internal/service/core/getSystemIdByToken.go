package core

import (
	"cactus/internal/storage/db"
	"context"
)

func (s *Service) GetTokenByPublicToken(ctx context.Context, token string) (db.Token, error) {
	dbToken, err := s.storage.GetTokenByPublicToken(ctx, token)
	return dbToken, err
}
