package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"cactus/apps/core/internal/service/core"
	"cactus/apps/core/internal/service/core/mocks"
)

// Структура для подготовки тестов
type testSetupGetStatus struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	service     *core.Service
}

func prepareTestGetStatus(t *testing.T) *testSetupGetStatus {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)

	service := core.New(mockStorage, nil, nil, nil, nil)

	return &testSetupGetStatus{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		service:     service,
	}
}

func TestGetStatus_Success(t *testing.T) {
	ts := prepareTestGetStatus(t)
	defer ts.ctrl.Finish()

	testUUID := uuid.New()
	expectedStatus := "processed"

	ts.mockStorage.EXPECT().GetStatusMessageByUUID(ts.ctx, testUUID).Return(expectedStatus, nil)

	result, err := ts.service.GetStatus(ts.ctx, testUUID.String())
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, result)
}

func TestGetStatus_Fail_InvalidUUID(t *testing.T) {
	ts := prepareTestGetStatus(t)
	defer ts.ctrl.Finish()

	invalidUUID := "invalid-uuid"

	result, err := ts.service.GetStatus(ts.ctx, invalidUUID)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestGetStatus_Fail_StorageError(t *testing.T) {
	ts := prepareTestGetStatus(t)
	defer ts.ctrl.Finish()

	testUUID := uuid.New()

	ts.mockStorage.EXPECT().GetStatusMessageByUUID(ts.ctx, testUUID).Return("", errors.New("ошибка базы данных"))

	result, err := ts.service.GetStatus(ts.ctx, testUUID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка базы данных")
	assert.Empty(t, result)
}
