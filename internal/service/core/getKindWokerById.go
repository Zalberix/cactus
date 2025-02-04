package core

import (
	"context"

	dto "cactus/internal/DTO"
)

func (s *Service) GetKindWokerByID(ctx context.Context, id int32) (dto.KindWorker, error) {
	kindWorker, err := s.storage.GetKindWokerByID(ctx, id)
	return dto.KindWorker(kindWorker), err
}
