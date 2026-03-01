package core

import (
	"context"
	"fmt"

	u "github.com/google/uuid"

	dto "cactus/apps/core/internal/DTO"
)

func (s *Service) GetFile(ctx context.Context, uuid string) (dto.GetFile, error) {
	UUID, err := u.Parse(uuid)
	if err != nil {
		return dto.GetFile{}, fmt.Errorf("uuid для файла не валидный: %v", err.Error())
	}

	path, err := s.storage.GetFilePathByUUID(ctx, UUID)
	if err != nil {
		return dto.GetFile{}, fmt.Errorf("файл по uuid не найден: %v", err.Error())
	}

	file, err := s.fileStorage.Get(ctx, path)
	if err != nil {
		return dto.GetFile{}, fmt.Errorf("%v", err.Error())
	}

	return dto.GetFile{
		File: file,
	}, nil
}
