package response

import (
	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
)

type GetMessagesResponse struct {
	Messages []dto.Message `json:"messages"`
}
