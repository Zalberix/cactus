package core_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	DTO "cactus/internal/DTO"
	configSchema "cactus/pkg/configschema"
	mocks_plugin "cactus/internal/plugin/mocks"
	"cactus/internal/service/core"
	"cactus/internal/service/core/mocks"
	"cactus/internal/storage/db"
)

type testSetupRegisterWorker struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	mockPlugins *mocks.MockPlugins
	service     *core.Service
}

func prepareTestRegisterWorker(t *testing.T) *testSetupRegisterWorker {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPlugins := mocks.NewMockPlugins(ctrl)

	service := core.New(mockStorage, nil, nil, mockPlugins, nil)

	return &testSetupRegisterWorker{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		mockPlugins: mockPlugins,
		service:     service,
	}
}

func TestRegisterWorker_Success(t *testing.T) {
	ts := prepareTestRegisterWorker(t)
	defer ts.ctrl.Finish()

	testUUID := uuid.New()
	testParams := core.RegisterWorkerParams{
		WorkerUUID:   testUUID,
		Kind:         "email",
		Type:         "worker_type",
		ConfigSchema: []configSchema.ConfigField{},
	}

	expectedKindWorker := db.KindWorker{ID: 1, Slug: "email"}
	expectedTypeWorker := db.TypeWorker{ID: 2, Slug: "worker_type"}
	expectedWorker := db.Worker{ID: 3, Uuid: testUUID, IDKindWorker: 1, IDTypeWorker: 2}

	mockPlugin := mocks_plugin.NewMockPlugin(ts.ctrl)

	ts.mockPlugins.EXPECT().Get(testParams.Kind).Return(mockPlugin, true)
	ts.mockStorage.EXPECT().SetContext(ts.ctx, gomock.Any()).Return(nil)
	ts.mockStorage.EXPECT().GetKindWorkerBySlug(ts.ctx, testParams.Kind).Return(expectedKindWorker, nil)
	ts.mockStorage.EXPECT().GetTypeWorkerBySlug(ts.ctx, testParams.Type).Return(expectedTypeWorker, nil)
	ts.mockStorage.EXPECT().GetWorkerByUUID(ts.ctx, testParams.WorkerUUID).Return(db.Worker{}, sql.ErrNoRows)
	ts.mockStorage.EXPECT().CreateWorker(ts.ctx, gomock.Any()).Return(expectedWorker, nil)
	ts.mockStorage.EXPECT().Commit().Return(nil)

	ts.mockStorage.EXPECT().Rollback().Times(1)

	result, err := ts.service.RegisterWorker(ts.ctx, testParams)
	assert.NoError(t, err)
	assert.True(t, result.Created)
	assert.Equal(t, expectedWorker.ID, result.ID)
}

func TestRegisterWorker_Fail_PluginNotFound(t *testing.T) {
	ts := prepareTestRegisterWorker(t)
	defer ts.ctrl.Finish()

	testParams := core.RegisterWorkerParams{
		WorkerUUID:   uuid.New(),
		Kind:         "unknown",
		Type:         "worker_type",
		ConfigSchema: []configSchema.ConfigField{},
	}

	ts.mockPlugins.EXPECT().Get(testParams.Kind).Return(nil, false)

	result, err := ts.service.RegisterWorker(ts.ctx, testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "воркеры с таким типом не включены или не поддерживаются")
	assert.Equal(t, DTO.RegisteWorker{}, result)
}

func TestRegisterWorker_Fail_StorageTransactionError(t *testing.T) {
	ts := prepareTestRegisterWorker(t)
	defer ts.ctrl.Finish()

	testParams := core.RegisterWorkerParams{
		WorkerUUID:   uuid.New(),
		Kind:         "email",
		Type:         "worker_type",
		ConfigSchema: []configSchema.ConfigField{},
	}

	mockPlugin := mocks_plugin.NewMockPlugin(ts.ctrl)

	ts.mockPlugins.EXPECT().Get(testParams.Kind).Return(mockPlugin, true)
	ts.mockStorage.EXPECT().SetContext(ts.ctx, gomock.Any()).Return(errors.New("транзакция не создана"))

	result, err := ts.service.RegisterWorker(ts.ctx, testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "невозможно зарегистрировать воркер")
	assert.Equal(t, DTO.RegisteWorker{}, result)
}

func TestRegisterWorker_Fail_GetKindWorkerError(t *testing.T) {
	ts := prepareTestRegisterWorker(t)
	defer ts.ctrl.Finish()

	testParams := core.RegisterWorkerParams{
		WorkerUUID:   uuid.New(),
		Kind:         "email",
		Type:         "worker_type",
		ConfigSchema: []configSchema.ConfigField{},
	}

	mockPlugin := mocks_plugin.NewMockPlugin(ts.ctrl)

	ts.mockPlugins.EXPECT().Get(testParams.Kind).Return(mockPlugin, true)
	ts.mockStorage.EXPECT().SetContext(ts.ctx, gomock.Any()).Return(nil)
	ts.mockStorage.EXPECT().
		GetKindWorkerBySlug(ts.ctx, testParams.Kind).
		Return(db.KindWorker{}, errors.New("ошибка при получении вида воркера"))

	// Ожидаем вызов Rollback(), так как транзакция неудачна
	ts.mockStorage.EXPECT().Rollback().Return(nil).Times(1)

	result, err := ts.service.RegisterWorker(ts.ctx, testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при получении вида воркер")
	assert.Equal(t, DTO.RegisteWorker{}, result)
}

func TestRegisterWorker_Fail_GetTypeWorkerError(t *testing.T) {
	ts := prepareTestRegisterWorker(t)
	defer ts.ctrl.Finish()

	testParams := core.RegisterWorkerParams{
		WorkerUUID:   uuid.New(),
		Kind:         "email",
		Type:         "worker_type",
		ConfigSchema: []configSchema.ConfigField{},
	}

	expectedKindWorker := db.KindWorker{ID: 1, Slug: "email"}

	mockPlugin := mocks_plugin.NewMockPlugin(ts.ctrl)

	ts.mockPlugins.EXPECT().Get(testParams.Kind).Return(mockPlugin, true)
	ts.mockStorage.EXPECT().SetContext(ts.ctx, gomock.Any()).Return(nil)
	ts.mockStorage.EXPECT().
		GetKindWorkerBySlug(ts.ctx, testParams.Kind).
		Return(expectedKindWorker, nil)

	ts.mockStorage.EXPECT().
		GetTypeWorkerBySlug(ts.ctx, testParams.Type).
		Return(db.TypeWorker{}, errors.New("ошибка при получении типа воркера"))

	ts.mockStorage.EXPECT().Rollback().Return(nil).Times(1)

	result, err := ts.service.RegisterWorker(ts.ctx, testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при получении типа воркер")
	assert.Equal(t, DTO.RegisteWorker{}, result)
}

func TestRegisterWorker_Fail_CreateWorkerError(t *testing.T) {
	ts := prepareTestRegisterWorker(t)
	defer ts.ctrl.Finish()

	testUUID := uuid.New()
	testParams := core.RegisterWorkerParams{
		WorkerUUID:   testUUID,
		Kind:         "email",
		Type:         "worker_type",
		ConfigSchema: []configSchema.ConfigField{},
	}

	expectedKindWorker := db.KindWorker{ID: 1, Slug: "email"}
	expectedTypeWorker := db.TypeWorker{ID: 2, Slug: "worker_type"}

	mockPlugin := mocks_plugin.NewMockPlugin(ts.ctrl)

	ts.mockPlugins.EXPECT().Get(testParams.Kind).Return(mockPlugin, true)
	ts.mockStorage.EXPECT().SetContext(ts.ctx, gomock.Any()).Return(nil)
	ts.mockStorage.EXPECT().GetKindWorkerBySlug(ts.ctx, testParams.Kind).Return(expectedKindWorker, nil)
	ts.mockStorage.EXPECT().
		GetTypeWorkerBySlug(ts.ctx, testParams.Type).Return(expectedTypeWorker, nil)
	ts.mockStorage.EXPECT().
		GetWorkerByUUID(ts.ctx, testParams.WorkerUUID).Return(db.Worker{}, sql.ErrNoRows)
	ts.mockStorage.EXPECT().
		CreateWorker(ts.ctx, gomock.Any()).Return(db.Worker{}, errors.New("ошибка при регистрации воркера"))

	ts.mockStorage.EXPECT().Rollback().Return(nil).Times(1)

	result, err := ts.service.RegisterWorker(ts.ctx, testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при регистрации воркера")
	assert.Equal(t, DTO.RegisteWorker{}, result)
}
