package email

import (
	"context"
)

func (s *Service) GetSystemIdByToken(ctx context.Context, token string) (int32, error) {
	system, err := s.storage.GetSystemIdByToken(ctx, token)
	return system, err
}
