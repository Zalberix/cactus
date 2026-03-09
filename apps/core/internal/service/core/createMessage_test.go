package core_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/libs/pipeline"
	pkgpipe "github.com/zalberix/cactus/libs/pipeline"
	mocks_plugin "github.com/zalberix/cactus/apps/core/internal/plugin/mocks"
	"github.com/zalberix/cactus/apps/core/internal/service/core"
	"github.com/zalberix/cactus/apps/core/internal/service/core/mocks"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

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
		TypeWorkerSlug: "email",
		Schema:         testSchema,
		SendLater:      &testTime,
		Files:          []core.SetFileParams{},
	}

	mockStorage.EXPECT().SetContext(ctx, gomock.Any()).Return(nil)
	mockStorage.EXPECT().
		GetPriorityBySystemId(ctx, testParams.IDSystem).
		Return(db.GetPriorityBySystemIdRow{Weight: 1}, nil)

	mockStorage.EXPECT().
		GetPriorityBySlug(ctx, testParams.PrioritySlug).
		Return(db.Priority{ID: 1, Weight: 5}, nil)

	mockStorage.EXPECT().
		GetTypeWorkerBySlug(ctx, testParams.TypeWorkerSlug).
		Return(db.TypeWorker{ID: 1}, nil)

	mockStorage.EXPECT().CreateManifest(ctx, gomock.Any()).Return(db.Manifest{ID: 1}, nil)

	mockStorage.EXPECT().CreateMessage(ctx, gomock.Any()).Return(db.Message{
		ID:       1,
		SystemID: testParams.IDSystem,
		ManifestID: 1,
		Uuid:     testUUID,
		Value:    schemaBytes,
		Priority: 6,
		SendAt: sql.NullTime{
			Time:  testTime,
			Valid: true,
		},
	}, nil)

	mockPlugin.EXPECT().ExtendPipeline([]pkgpipe.Step{}).Return([]pkgpipe.Step{{Name: "Step1"}}, nil)
	mockPipelineService.EXPECT().CreatePipelineTX(ctx, mockStorage, gomock.Any()).Return([]dto.Pipeline{
		{
			ID:         1,
			PipelineID: 1,
			WorkerID: sql.NullInt32{
				Valid: false,
			},
			Step: 1,
			TimeStart: sql.NullTime{
				Valid: false,
			},
			TimeEnd: sql.NullTime{
				Valid: false,
			},
			PipelineStepStatusID: 1,
		},
	}, nil)

	// s.AddMessageToQueue
	nameQueue := "messages:email:test_kind:w-6"
	mockBroker.EXPECT().EnsureStreamGroup(ctx, nameQueue, "reader").Return(nil)
	mockStorage.EXPECT().GetSystemById(ctx, gomock.Any()).Return(db.System{Name: "test"}, nil)
	mockBroker.EXPECT().
		AddMessageToQueue(
			ctx, nameQueue, gomock.Any(),
			pipeline.SystemInQueue{Name: "test"},
			pipeline.PipelineInQueue{Step: 1},
		).
		Return(nil)

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
	mockStorage.EXPECT().
		GetPriorityBySystemId(ctx, gomock.Any()).
		Return(db.GetPriorityBySystemIdRow{}, errors.New("ошибка получения приоритета системы"))

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
	mockStorage.EXPECT().
		GetPriorityBySystemId(ctx, gomock.Any()).
		Return(db.GetPriorityBySystemIdRow{Weight: 1}, nil)

	mockStorage.EXPECT().
		GetPriorityBySlug(ctx, gomock.Any()).
		Return(db.Priority{}, errors.New("ошибка получения приоритета по slug"))

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
	mockStorage.EXPECT().
		GetPriorityBySystemId(ctx, gomock.Any()).
		Return(db.GetPriorityBySystemIdRow{Weight: 10}, nil)

	mockStorage.EXPECT().
		GetPriorityBySlug(ctx, gomock.Any()).
		Return(db.Priority{ID: 1, Weight: 5}, nil)

	mockStorage.EXPECT().
		GetTypeWorkerBySlug(ctx, gomock.Any()).
		Return(db.TypeWorker{}, errors.New("ошибка получения типа воркера по slug"))

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
	mockStorage.EXPECT().CreateManifest(ctx, gomock.Any()).Return(db.Manifest{ID: 1}, nil)
	mockStorage.EXPECT().CreateMessage(ctx, gomock.Any()).Return(db.Message{}, errors.New("ошибка создания сообщения"))

	mockStorage.EXPECT().Rollback().Times(1)

	service := core.New(mockStorage, nil, nil, nil, nil)

	_, err := service.CreateMessage(ctx, core.CreateMessageParams{}, mockPipelineService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания сообщения")
}
