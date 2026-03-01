package core

import (
	"context"
	"fmt"

	dto "cactus/apps/core/internal/DTO"
)

func (s *Service) GetTypeWorkers(ctx context.Context) ([]dto.TypeWorker, error) {
	typeWorkersModel, err := s.storage.GetTypeWorkers(ctx)
	if err != nil {
		return []dto.TypeWorker{}, fmt.Errorf("%v", err.Error())
	}

	typeWorkers := make([]dto.TypeWorker, 0, len(typeWorkersModel))

	for _, typeWorker := range typeWorkersModel {
		typeWorkers = append(typeWorkers, dto.TypeWorker{
			TypeWorker: typeWorker,
		})
	}

	return typeWorkers, err
}
