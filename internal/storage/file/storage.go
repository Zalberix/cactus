package file

import (
	"github.com/minio/minio-go/v7"
)

type FileStorage struct {
	Storage *minio.Client
}

func New(minio *minio.Client) *FileStorage {
	return &FileStorage{
		Storage: minio,
	}
}
