package email

import (
	"github.com/google/uuid"
)

type StatusRequest struct {
	UUID uuid.UUID `json:"uuid"`
}

type RenderRequest struct {
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
	Theme   string            `json:"theme"`
}

type LinkRequest struct {
	UUID uuid.UUID `json:"uuid"`
}
