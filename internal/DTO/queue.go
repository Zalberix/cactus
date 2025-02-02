package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SystemValueInMessageQueue struct {
	Name string `json:"name"`
}
type FileInMessageValueInMessageQueue struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}
type MessageValueInMessageQueue struct {
	ID        int32                              `json:"id"`
	UUID      uuid.UUID                          `json:"uuid"`
	Value     json.RawMessage                    `json:"value"`
	SendLater *time.Time                         `json:"send_later,omitempty"`
	CreateAt  time.Time                          `json:"create_at"`
	Files     []FileInMessageValueInMessageQueue `json:"files"`
}
