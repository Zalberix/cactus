package email

import (
	dto "cactus/internal/DTO"
	"cactus/internal/schema"
)

type GetEmailsResponse struct {
	Emails []dto.Message[schema.Email] `json:"emails"`
}
