package email

import (
	"github.com/google/uuid"
)

// TODO: нужен ли он?
type LinkRequest struct {
	UUID uuid.UUID `json:"uuid"`
}
