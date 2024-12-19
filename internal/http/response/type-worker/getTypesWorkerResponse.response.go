package type_worker

import (
	dto "cactus/internal/DTO"
)

type GetTypesWorkerResponse struct {
	TypesWorker []dto.TypeWorker `json:"types_worker"`
}
