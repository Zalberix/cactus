package core

import (
	dto "cactus/internal/DTO"
	"context"
	"fmt"
)

func (s *Service) GetTypeWorkers(ctx context.Context) ([]dto.TypeWorker, error) {
	typeWorkersModel, err := s.storage.GetTypeWorkers(ctx)
	if err != nil {
		return []dto.TypeWorker{}, fmt.Errorf("%v", err.Error())
	}

	var typeWorkers []dto.TypeWorker

	for _, typeWorker := range typeWorkersModel {
		typeWorkers = append(typeWorkers, dto.TypeWorker{
			TypeWorker: typeWorker,
		})
	}

	return typeWorkers, err
}
