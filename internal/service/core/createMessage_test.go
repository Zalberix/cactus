package core_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	dto "cactus/internal/DTO"
	pkgpipe "cactus/internal/pkg/pipeline"
	mocks_plugin "cactus/internal/plugin/mocks"
	"cactus/internal/service/core"
	"cactus/internal/service/core/mocks"
	"cactus/internal/storage/db"
)

type MockService struct {
	core.Service
}

func (m *MockService) AddMessageToQueue(ctx context.Context, params core.AddMessageToQueueParams) error {
	return nil
}

func (m *MockService) SetFileTX(ctx context.Context, storage mocks.MockStorage, params core.SetFileParams) (dto.SetFile, error) {
	return dto.SetFile{}, nil
}

func TestCreateMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockBroker := mocks.NewMockBroker(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	mockPlugins := mocks.NewMockPlugins(ctrl)
	mockPlugin := mocks_plugin.NewMockPlugin(ctrl)

	testUUID := uuid.New()
	testTime := time.Now()

	testSchema := map[string]interface{}{
		"key": "value",
	}
	schemaBytes, _ := json.Marshal(testSchema)

	testParams := core.CreateMessageParams{
		Plugin:         mockPlugin,
		KindWorkerSlug: "test_kind",
		IDSystem:       1,
		PrioritySlug:   "low",
		ChanelSlug:     "email",
		Schema:         testSchema,
		SendLater:      &testTime,
		Files:          []core.SetFileParams{},
	}

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, testParams.IDSystem).Return(db.GetPriorityBySystemIdRow{Weight: 1}, nil)
	mockStorage.EXPECT().GetPriorityBySlug(ctx, testParams.PrioritySlug).Return(db.Priority{ID: 1, Weight: 5}, nil)
	mockStorage.EXPECT().GetTypeWorkerBySlug(ctx, testParams.ChanelSlug).Return(db.TypeWorker{ID: 1}, nil)

	// s.SetFileTX
	// mockFileStorage.EXPECT().Save(ctx, gomock.Any(), gomock.Any()).Return("", nil)
	// mockStorage.EXPECT().CreateFile(ctx, db.CreateFileParams{}).Return(db.File{
	// 	Uuid:  uuid.New(),
	// 	Title: "Заголовок",
	// 	Ext:   "png",
	// }, nil)

	mockStorage.EXPECT().CreateMessage(ctx, gomock.Any()).Return(db.Message{
		ID:           1,
		IDTypeWorker: 1,
		IDSystem:     testParams.IDSystem,
		Uuid:         testUUID,
		Value:        schemaBytes,
		IDPriority:   1,
		SendLater: sql.NullTime{
			Time:  testTime,
			Valid: true,
		},
	}, nil)

	mockPlugin.EXPECT().ExtendPipeline([]pkgpipe.Step{}).Return([]pkgpipe.Step{{Name: "Step1"}}, nil)
	mockPipelineService.EXPECT().CreatePipelineTX(ctx, mockStorage, gomock.Any()).Return([]dto.Pipeline{
		{
			ID:        1,
			IDMessage: 1,
			IDWorker: sql.NullInt32{
				Valid: false,
			},
			Status: "wait",
			Step:   1,
			Name:   "Нзавние шага",
			TimeStart: sql.NullTime{
				Valid: false,
			},
			TimeEnd: sql.NullTime{
				Valid: false,
			},
		},
	}, nil)

	// s.AddMessageToQueue
	nameQueue := "messages:email:test_kind:w-6"
	mockBroker.EXPECT().EnsureStreamGroup(ctx, nameQueue, "reader").Return(nil)
	mockStorage.EXPECT().GetSystemById(ctx, gomock.Any()).Return(db.System{Name: "test"}, nil)
	mockBroker.EXPECT().AddMessageToQueue(ctx, nameQueue, gomock.Any(), dto.SystemValueInMessageQueue{Name: "test"}, dto.PipelineValueInMessageQueue{Step: 1}).Return(nil)

	mockStorage.EXPECT().Commit().Return(nil)
	mockStorage.EXPECT().Rollback().Times(1)

	// File

	service := core.New(mockStorage, mockBroker, mockFileStorage, mockPlugins, nil)

	result, err := service.CreateMessage(ctx, testParams, mockPipelineService)

	assert.NoError(t, err)
	assert.Equal(t, testUUID, result.Message.Message.Uuid)
	assert.Equal(t, testSchema, result.Message.Value)

	var files []dto.SetFile

	assert.Equal(t, &files, result.Files)
}

func TestCreateMessage_SuccessWithFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockBroker := mocks.NewMockBroker(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	mockPlugins := mocks.NewMockPlugins(ctrl)
	mockPlugin := mocks_plugin.NewMockPlugin(ctrl)

	testUUID := uuid.New()
	testTime := time.Now()

	testSchema := map[string]interface{}{
		"key": "value",
	}
	schemaBytes, _ := json.Marshal(testSchema)

	file, _, _ := os.Pipe()
	ext := *mimetype.Lookup("image/png")

	testParams := core.CreateMessageParams{
		Plugin:         mockPlugin,
		KindWorkerSlug: "test_kind",
		IDSystem:       1,
		PrioritySlug:   "low",
		ChanelSlug:     "email",
		Schema:         testSchema,
		SendLater:      &testTime,
		Files: []core.SetFileParams{
			{
				Title:     "Заголовок1",
				IDMessage: int32(1),
				File:      file,
				Ext:       ext,
			},
		},
	}

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, testParams.IDSystem).Return(db.GetPriorityBySystemIdRow{Weight: 1}, nil)
	mockStorage.EXPECT().GetPriorityBySlug(ctx, testParams.PrioritySlug).Return(db.Priority{ID: 1, Weight: 5}, nil)
	mockStorage.EXPECT().GetTypeWorkerBySlug(ctx, testParams.ChanelSlug).Return(db.TypeWorker{ID: 1}, nil)

	// s.SetFileTX
	mockFileStorage.EXPECT().Save(ctx, file, ext).Return("", nil)
	mockStorage.EXPECT().CreateFile(ctx, db.CreateFileParams{
		Ext: "png",
	}).Return(db.File{
		Uuid:  uuid.New(),
		Title: "Заголовок",
		Ext:   "png",
	}, nil)

	mockStorage.EXPECT().CreateMessage(ctx, gomock.Any()).Return(db.Message{
		ID:           1,
		IDTypeWorker: 1,
		IDSystem:     testParams.IDSystem,
		Uuid:         testUUID,
		Value:        schemaBytes,
		IDPriority:   1,
		SendLater: sql.NullTime{
			Time:  testTime,
			Valid: true,
		},
	}, nil)

	mockPlugin.EXPECT().ExtendPipeline([]pkgpipe.Step{}).Return([]pkgpipe.Step{{Name: "Step1"}}, nil)
	mockPipelineService.EXPECT().CreatePipelineTX(ctx, mockStorage, gomock.Any()).Return([]dto.Pipeline{
		{
			ID:        1,
			IDMessage: 1,
			IDWorker: sql.NullInt32{
				Valid: false,
			},
			Status: "wait",
			Step:   1,
			Name:   "Нзавние шага",
			TimeStart: sql.NullTime{
				Valid: false,
			},
			TimeEnd: sql.NullTime{
				Valid: false,
			},
		},
	}, nil)

	// s.AddMessageToQueue
	nameQueue := "messages:email:test_kind:w-6"
	mockBroker.EXPECT().EnsureStreamGroup(ctx, nameQueue, "reader").Return(nil)
	mockStorage.EXPECT().GetSystemById(ctx, gomock.Any()).Return(db.System{Name: "test"}, nil)
	mockBroker.EXPECT().AddMessageToQueue(ctx, nameQueue, gomock.Any(), dto.SystemValueInMessageQueue{Name: "test"}, dto.PipelineValueInMessageQueue{Step: 1}).Return(nil)

	mockStorage.EXPECT().Commit().Return(nil)
	mockStorage.EXPECT().Rollback().Times(1)

	// File

	service := core.New(mockStorage, mockBroker, mockFileStorage, mockPlugins, nil)

	result, err := service.CreateMessage(ctx, testParams, mockPipelineService)

	assert.NoError(t, err)
	assert.Equal(t, testUUID, result.Message.Message.Uuid)
	assert.Equal(t, testSchema, result.Message.Value)

	var files []dto.SetFile

	assert.Equal(t, &files, result.Files)
}

func TestCreateMessage_Fail_SetContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(errors.New("ошибка создания контекста"))

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "невозможно создать транзакцию для сообщения")
}

func TestCreateMessage_Fail_GetPriorityBySystemId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, gomock.Any()).Return(db.GetPriorityBySystemIdRow{}, errors.New("ошибка получения приоритета системы"))

	mockStorage.EXPECT().Rollback().Times(1)

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения приоритета системы")
}

func TestCreateMessage_Fail_GetPriorityBySlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, gomock.Any()).Return(db.GetPriorityBySystemIdRow{Weight: 1}, nil)
	mockStorage.EXPECT().GetPriorityBySlug(ctx, gomock.Any()).Return(db.Priority{}, errors.New("ошибка получения приоритета по slug"))

	mockStorage.EXPECT().Rollback().Times(1)

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения приоритета по slug")
}

func TestCreateMessage_Fail_GetTypeWorkerBySlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, gomock.Any()).Return(db.GetPriorityBySystemIdRow{Weight: 10}, nil)
	mockStorage.EXPECT().GetPriorityBySlug(ctx, gomock.Any()).Return(db.Priority{ID: 1, Weight: 5}, nil)
	mockStorage.EXPECT().GetTypeWorkerBySlug(ctx, gomock.Any()).Return(db.TypeWorker{}, errors.New("ошибка получения типа воркера по slug"))

	mockStorage.EXPECT().Rollback().Times(1)

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения типа воркера по slug")
}

func TestCreateMessage_Fail_CreateMessageDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, gomock.Any()).Return(db.GetPriorityBySystemIdRow{Weight: 10}, nil)
	mockStorage.EXPECT().GetPriorityBySlug(ctx, gomock.Any()).Return(db.Priority{ID: 1, Weight: 5}, nil)
	mockStorage.EXPECT().GetTypeWorkerBySlug(ctx, gomock.Any()).Return(db.TypeWorker{ID: 1}, nil)
	mockStorage.EXPECT().CreateMessage(ctx, gomock.Any()).Return(db.Message{}, errors.New("ошибка создания сообщения:"))

	mockStorage.EXPECT().Rollback().Times(1)

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания сообщения:")
}

func TestCreateMessage_Fail_PluginExtendPipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPipelineService := mocks.NewMockPipelineService(ctrl)
	mockPlugin := mocks_plugin.NewMockPlugin(ctrl)

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().GetPriorityBySystemId(ctx, gomock.Any()).Return(db.GetPriorityBySystemIdRow{Weight: 10}, nil)
	mockStorage.EXPECT().GetPriorityBySlug(ctx, gomock.Any()).Return(db.Priority{ID: 1, Weight: 5}, nil)
	mockStorage.EXPECT().GetTypeWorkerBySlug(ctx, gomock.Any()).Return(db.TypeWorker{ID: 1}, nil)
	mockStorage.EXPECT().CreateMessage(ctx, gomock.Any()).Return(db.Message{ID: 1}, nil)

	mockPlugin.EXPECT().ExtendPipeline(gomock.Any()).Return(nil, errors.New("ошибка pipeline"))

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{Plugin: mockPlugin}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при расширении pipepline")
}
