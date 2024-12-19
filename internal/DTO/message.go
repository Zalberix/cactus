package dto

import (
	"cactus/internal/storage/db"
)

type Message[T any] struct {
	db.Message
	Value T `json:"value"`
}
