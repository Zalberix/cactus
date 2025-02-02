package core

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	dto "cactus/internal/DTO"
	"cactus/internal/storage/db"
)

type SetFileParams struct {
	IDMessage int32
	File      multipart.File
	Title     string
	Ext       mimetype.MIME
}

func (s *Service) SetFile(ctx context.Context, params SetFileParams) (dto.SetFile, error) {
	return s.SetFileTX(ctx, nil, params)
}

func (s *Service) SetFileTX(ctx context.Context, tx *sql.Tx, params SetFileParams) (dto.SetFile, error) {
	storage := s.storage
	if tx != nil {
		storage = storage.WithTx(tx)
	}

	path, err := s.fileStorage.Save(ctx, params.File, params.Ext)
	if err != nil {
		return dto.SetFile{}, fmt.Errorf("ошибка сохранения файла %v", err.Error())
	}

	file, err := storage.CreateFile(ctx, db.CreateFileParams{
		IDMessage: params.IDMessage,
		Title:     params.Title,
		Path:      path,
		Ext:       params.Ext.Extension()[1:],
		Uuid:      uuid.New(),
	})
	if err != nil {
		return dto.SetFile{}, err
	}

	return dto.SetFile{
		UUID:  file.Uuid,
		Title: file.Title,
		Ext:   file.Ext,
	}, nil
}
