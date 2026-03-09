package response

import (
	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
)

type GetTypesWorkerResponse struct {
	TypesWorker []dto.TypeWorker `json:"types_worker"`
}
