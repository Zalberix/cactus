package core

import (
	"context"
	"net/http"

	dto "cactus/apps/core/internal/DTO"
	"cactus/apps/core/internal/http/response"
)

type getTypesWorkerService interface {
	GetTypeWorkers(ctx context.Context) ([]dto.TypeWorker, error)
}

func GetTypesWorker(s getTypesWorkerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO обработать ошибку
		typeWorker, _ := s.GetTypeWorkers(r.Context())
		response.OKJSON(w, response.GetTypesWorkerResponse{
			TypesWorker: typeWorker,
		})
	}
}
