package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/internal/service/core"
	"github.com/zalberix/cactus/apps/core/internal/service/core/mocks"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

type testSetupGetKindWorker struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	service     *core.Service
}

func prepareTestGetKindWorker(t *testing.T) *testSetupGetKindWorker {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)

	service := core.New(mockStorage, nil, nil, nil, nil)

	return &testSetupGetKindWorker{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		service:     service,
	}
}

func TestGetKindWorkerByID_Success(t *testing.T) {
	ts := prepareTestGetKindWorker(t)
	defer ts.ctrl.Finish()

	expectedKindWorker := db.KindWorker{
		ID:   1,
		Name: "Test Worker",
	}

	ts.mockStorage.EXPECT().GetKindWokerByID(ts.ctx, int32(1)).Return(expectedKindWorker, nil)

	result, err := ts.service.GetKindWokerByID(ts.ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, dto.KindWorker(expectedKindWorker), result)
}

func TestGetKindWorkerByID_Fail_NotFound(t *testing.T) {
	ts := prepareTestGetKindWorker(t)
	defer ts.ctrl.Finish()

	ts.mockStorage.EXPECT().GetKindWokerByID(ts.ctx, int32(1)).Return(db.KindWorker{}, errors.New("не найдено"))

	result, err := ts.service.GetKindWokerByID(ts.ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено")
	assert.Equal(t, dto.KindWorker{}, result)
}

func TestGetKindWorkerByID_Fail_StorageError(t *testing.T) {
	ts := prepareTestGetKindWorker(t)
	defer ts.ctrl.Finish()

	ts.mockStorage.EXPECT().GetKindWokerByID(ts.ctx, int32(1)).Return(db.KindWorker{}, errors.New("ошибка базы данных"))

	result, err := ts.service.GetKindWokerByID(ts.ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка базы данных")
	assert.Equal(t, dto.KindWorker{}, result)
}
