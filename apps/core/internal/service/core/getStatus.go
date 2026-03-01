package core

import (
	"context"

	u "github.com/google/uuid"
)

func (s *Service) GetStatus(ctx context.Context, uuid string) (string, error) {
	uuidO, err := u.Parse(uuid)
	if err != nil {
		return "", err
	}
	str, err := s.storage.GetStatusMessageByUUID(ctx, uuidO)
	return str, err
}
