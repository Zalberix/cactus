package core_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"cactus/internal/service/core"
	"cactus/internal/service/core/mocks"
)

func TestGetFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	testUUID := uuid.New()
	testPath := "test/path/to/file"

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	mockFile, _, _ := os.Pipe()

	mockStorage.EXPECT().GetFilePathByUUID(ctx, testUUID).Return(testPath, nil)
	mockFileStorage.EXPECT().Get(ctx, testPath).Return(mockFile, nil)

	service := core.New(mockStorage, nil, mockFileStorage, nil, nil)

	result, err := service.GetFile(ctx, testUUID.String())

	assert.NoError(t, err)
	assert.Equal(t, mockFile, result.File)
}

func TestGetFile_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	invalidUUID := "invalid-uuid"

	service := core.New(nil, nil, nil, nil, nil)

	_, err := service.GetFile(ctx, invalidUUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid для файла не валидный")
}

func TestGetFile_FileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	testUUID := uuid.New()

	mockStorage := mocks.NewMockStorage(ctrl)

	mockStorage.EXPECT().GetFilePathByUUID(ctx, testUUID).Return("", errors.New("file not found"))

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.GetFile(ctx, testUUID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "файл по uuid не найден")
}

func TestGetFile_FileReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	testUUID := uuid.New()
	testPath := "test/path/to/file"

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)

	mockStorage.EXPECT().GetFilePathByUUID(ctx, testUUID).Return(testPath, nil)
	mockFileStorage.EXPECT().Get(ctx, testPath).Return(nil, errors.New("read error"))

	service := core.New(mockStorage, nil, mockFileStorage, nil, nil)

	_, err := service.GetFile(ctx, testUUID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}
