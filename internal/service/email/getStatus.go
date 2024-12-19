package email

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) GetStatus(ctx context.Context, uuidRow string) (string, error) {
	uuid, err := uuid.Parse(uuidRow)
	if err != nil {
		return "", err
	}
	str, err := s.storage.GetStatusMessageByUUID(ctx, uuid)
	return str, err
}
