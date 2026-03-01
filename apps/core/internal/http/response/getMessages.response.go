package response

import (
	dto "cactus/apps/core/internal/DTO"
)

type GetMessagesResponse struct {
	Messages []dto.Message `json:"messages"`
}
