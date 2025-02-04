package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"cactus/internal/service/core"
	"cactus/internal/service/core/mocks"
	"cactus/internal/storage/db"
)

type testSetupGetTypeWorkers struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	service     *core.Service
}

func prepareTestGetTypeWorkers(t *testing.T) *testSetupGetTypeWorkers {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)

	service := core.New(mockStorage, nil, nil, nil, nil)

	return &testSetupGetTypeWorkers{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		service:     service,
	}
}

func TestGetTypeWorkers_Success(t *testing.T) {
	ts := prepareTestGetTypeWorkers(t)
	defer ts.ctrl.Finish()

	expectedTypeWorkers := []db.TypeWorker{
		{ID: 1, Name: "Worker1"},
		{ID: 2, Name: "Worker2"},
	}

	ts.mockStorage.EXPECT().GetTypeWorkers(ts.ctx).Return(expectedTypeWorkers, nil)

	result, err := ts.service.GetTypeWorkers(ts.ctx)
	assert.NoError(t, err)
	assert.Len(t, result, len(expectedTypeWorkers))
	assert.Equal(t, expectedTypeWorkers[0].ID, result[0].ID)
	assert.Equal(t, expectedTypeWorkers[1].ID, result[1].ID)
}

func TestGetTypeWorkers_Success_EmptyList(t *testing.T) {
	ts := prepareTestGetTypeWorkers(t)
	defer ts.ctrl.Finish()

	ts.mockStorage.EXPECT().GetTypeWorkers(ts.ctx).Return([]db.TypeWorker{}, nil)

	result, err := ts.service.GetTypeWorkers(ts.ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestGetTypeWorkers_Fail_StorageError(t *testing.T) {
	ts := prepareTestGetTypeWorkers(t)
	defer ts.ctrl.Finish()

	ts.mockStorage.EXPECT().GetTypeWorkers(ts.ctx).Return(nil, errors.New("ошибка базы данных"))

	result, err := ts.service.GetTypeWorkers(ts.ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка базы данных")
	assert.Empty(t, result)
}

func TestGetTypeWorkers_Fail_NilReturn(t *testing.T) {
	ts := prepareTestGetTypeWorkers(t)
	defer ts.ctrl.Finish()

	ts.mockStorage.EXPECT().GetTypeWorkers(ts.ctx).Return(nil, nil)

	result, err := ts.service.GetTypeWorkers(ts.ctx)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetTypeWorkers_Fail_CorruptedData(t *testing.T) {
	ts := prepareTestGetTypeWorkers(t)
	defer ts.ctrl.Finish()

	expectedTypeWorkers := []db.TypeWorker{
		{ID: 1, Name: ""},
		{ID: 2, Name: ""},
	}

	ts.mockStorage.EXPECT().GetTypeWorkers(ts.ctx).Return(expectedTypeWorkers, nil)

	result, err := ts.service.GetTypeWorkers(ts.ctx)
	assert.NoError(t, err)
	assert.Len(t, result, len(expectedTypeWorkers))
	assert.Equal(t, expectedTypeWorkers[0].ID, result[0].ID)
	assert.Equal(t, expectedTypeWorkers[1].ID, result[1].ID)
}
