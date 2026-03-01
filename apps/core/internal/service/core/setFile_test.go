package core_test

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"testing"

	"github.com/gabriel-vasile/mimetype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	dto "cactus/apps/core/internal/DTO"
	"cactus/apps/core/internal/service/core"
	"cactus/apps/core/internal/service/core/mocks"
	"cactus/apps/core/storage/db"
)

type testSetupSetFile struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	mockFile    *mocks.MockFileStorage
	service     *core.Service
}

func prepareTestSetFile(t *testing.T) *testSetupSetFile {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)

	service := core.New(mockStorage, nil, mockFile, nil, nil)

	return &testSetupSetFile{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		mockFile:    mockFile,
		service:     service,
	}
}

func TestSetFile_Success(t *testing.T) {
	ts := prepareTestSetFile(t)
	defer ts.ctrl.Finish()

	file, _ := os.Open("test_file.txt")
	defer file.Close()

	params := core.SetFileParams{
		IDMessage: 1,
		File:      file,
		Title:     "test_file",
		Ext:       *mimetype.Lookup("text/plain"),
	}

	const expectedPath = "/path/to/file"
	expectedFile := db.File{
		Title:     "test_file",
		Ext:       "txt",
		MessageID: 1,
		Url:       expectedPath,
	}

	ts.mockFile.EXPECT().Save(ts.ctx, gomock.Any(), gomock.AssignableToTypeOf(mimetype.MIME{})).Return(expectedPath, nil)

	ts.mockStorage.EXPECT().CreateFile(ts.ctx, gomock.Any()).Return(expectedFile, nil)

	result, err := ts.service.SetFile(ts.ctx, params)
	assert.NoError(t, err)
	assert.NotZero(t, result.UUID)
	assert.Equal(t, expectedFile.Title, result.Title)
	assert.Equal(t, expectedFile.Ext, result.Ext)
}

func TestSetFile_Fail_FileSaveError(t *testing.T) {
	ts := prepareTestSetFile(t)
	defer ts.ctrl.Finish()

	file, _ := os.Open("test_file.txt")
	defer file.Close()

	params := core.SetFileParams{
		IDMessage: 1,
		File:      file,
		Title:     "test_file",
		Ext:       *mimetype.Lookup("text/plain"),
	}

	ts.mockFile.EXPECT().
		Save(ts.ctx, gomock.Any(), gomock.AssignableToTypeOf(mimetype.MIME{})).
		DoAndReturn(func(_ context.Context, _ multipart.File, _ mimetype.MIME) (string, error) {
			return "", errors.New("ошибка сохранения файла")
		})

	result, err := ts.service.SetFile(ts.ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка сохранения файла")
	assert.Equal(t, dto.SetFile{}, result)
}

func TestSetFile_Fail_CreateFileError(t *testing.T) {
	ts := prepareTestSetFile(t)
	defer ts.ctrl.Finish()

	file, _ := os.Open("test_file.txt")
	defer file.Close()

	params := core.SetFileParams{
		IDMessage: 1,
		File:      file,
		Title:     "test_file",
		Ext:       *mimetype.Lookup("text/plain"),
	}

	const expectedPath = "/path/to/file"

	ts.mockFile.EXPECT().Save(ts.ctx, gomock.Any(), gomock.AssignableToTypeOf(mimetype.MIME{})).Return(expectedPath, nil)

	ts.mockStorage.EXPECT().CreateFile(ts.ctx, gomock.Any()).Return(db.File{}, errors.New("ошибка создания файла в БД"))

	result, err := ts.service.SetFile(ts.ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания файла в БД")
	assert.Equal(t, dto.SetFile{}, result)
}

func TestSetFileTX_Success(t *testing.T) {
	ts := prepareTestSetFile(t)
	defer ts.ctrl.Finish()

	file, _ := os.Open("test_file.txt")
	defer file.Close()

	params := core.SetFileParams{
		IDMessage: 1,
		File:      file,
		Title:     "test_file",
		Ext:       *mimetype.Lookup("text/plain"),
	}

	expectedPath := "/path/to/file"
	expectedFile := db.File{
		Title:     "test_file",
		Ext:       "txt",
		MessageID: 1,
		Url:       expectedPath,
	}

	ts.mockFile.EXPECT().Save(ts.ctx, gomock.Any(), gomock.AssignableToTypeOf(mimetype.MIME{})).Return(expectedPath, nil)

	ts.mockStorage.EXPECT().CreateFile(ts.ctx, gomock.Any()).Return(expectedFile, nil)

	result, err := ts.service.SetFileTX(ts.ctx, ts.mockStorage, params)
	assert.NoError(t, err)
	assert.NotZero(t, result.UUID)
	assert.Equal(t, expectedFile.Title, result.Title)
	assert.Equal(t, expectedFile.Ext, result.Ext)
}

func TestSetFileTX_Fail_CreateFileError(t *testing.T) {
	ts := prepareTestSetFile(t)
	defer ts.ctrl.Finish()

	file, _ := os.Open("test_file.txt")
	defer file.Close()

	params := core.SetFileParams{
		IDMessage: 1,
		File:      file,
		Title:     "test_file",
		Ext:       *mimetype.Lookup("text/plain"),
	}

	expectedPath := "/path/to/file"

	ts.mockFile.EXPECT().Save(ts.ctx, gomock.Any(), gomock.AssignableToTypeOf(mimetype.MIME{})).Return(expectedPath, nil)

	ts.mockStorage.EXPECT().CreateFile(ts.ctx, gomock.Any()).Return(db.File{}, errors.New("ошибка создания файла в БД"))

	result, err := ts.service.SetFileTX(ts.ctx, ts.mockStorage, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания файла в БД")
	assert.Equal(t, dto.SetFile{}, result)
}
