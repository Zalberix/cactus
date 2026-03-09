package core

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

type SetFileParams struct {
	IDMessage int32
	File      multipart.File
	Title     string
	Ext       mimetype.MIME
}

func (s *Service) SetFile(ctx context.Context, params SetFileParams) (dto.SetFile, error) {
	return s.setFile(ctx, params, s.storage)
}

func (s *Service) SetFileTX(ctx context.Context, storage Storage, params SetFileParams) (dto.SetFile, error) {
	return s.setFile(ctx, params, storage)
}

func (s *Service) setFile(
	ctx context.Context,
	params SetFileParams,
	storage Storage,
) (dto.SetFile, error) {
	path, err := s.fileStorage.Save(ctx, params.File, params.Ext)
	if err != nil {
		return dto.SetFile{}, fmt.Errorf("ошибка сохранения файла %v", err.Error())
	}

	fileUUID := uuid.New()
	ext := params.Ext.Extension()[1:]
	file, err := storage.CreateFile(ctx, db.CreateFileParams{
		MessageID: params.IDMessage,
		Title:     params.Title,
		Name:      fmt.Sprintf("%s.%s", params.Title, ext),
		Ext:       ext,
		Url:       path,
	})
	if err != nil {
		return dto.SetFile{}, err
	}

	return dto.SetFile{
		UUID:  fileUUID,
		Title: file.Title,
		Ext:   file.Ext,
	}, nil
}
