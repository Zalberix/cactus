package dto

import (
	"os"

	"github.com/google/uuid"
)

type SetFile struct {
	UUID  uuid.UUID
	Title string
	Ext   string
}

type GetFile struct {
	File *os.File
}
