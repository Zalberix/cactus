package dto

import (
	"cactus/apps/core/storage/db"
)

type Message struct {
	db.Message
	Value any `json:"value"`
}

type CreateMessage struct {
	Message
	Files *[]SetFile `json:"files"`
}
