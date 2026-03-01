package response

import (
	dto "cactus/apps/core/internal/DTO"
)

type GetTypesWorkerResponse struct {
	TypesWorker []dto.TypeWorker `json:"types_worker"`
}
