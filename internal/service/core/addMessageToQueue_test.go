package core_test

import (
	"cactus/internal/server/meta"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	dto "cactus/internal/DTO"
	"cactus/internal/service/core"
	"cactus/internal/service/core/mocks"
	"cactus/internal/storage/db"
)

type testSetup struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	mockBroker  *mocks.MockBroker
	service     *core.Service
	testParams  core.AddMessageToQueueParams
	queueName   string
}

func prepareTest(t *testing.T) *testSetup {
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockBroker := mocks.NewMockBroker(ctrl)
	mockMeta := &meta.ServerMeta{HostName: "localhost", Port: "8080"}

	service := core.New(mockStorage, mockBroker, nil, nil, mockMeta)

	testUUID := uuid.New()
	testParams := core.AddMessageToQueueParams{
		Message: db.Message{
			ID:       1,
			Uuid:     testUUID,
			IDSystem: 1,
		},
		SlugKindWorker:        "test_kind",
		SlugTypeWorker:        "email",
		WeightPriorityMessage: 5,
		Step:                  1,
	}

	queueName := "messages:email:test_kind:w-5"

	return &testSetup{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		mockBroker:  mockBroker,
		service:     service,
		testParams:  testParams,
		queueName:   queueName,
	}
}

func TestAddMessageToQueue_Success(t *testing.T) {
	ts := prepareTest(t)
	defer ts.ctrl.Finish()

	ts.mockBroker.EXPECT().EnsureStreamGroup(ts.ctx, ts.queueName, "reader").Return(nil)
	ts.mockStorage.EXPECT().GetSystemById(ts.ctx, ts.testParams.IDSystem).Return(db.System{Name: "TestSystem"}, nil)
	ts.mockBroker.EXPECT().AddMessageToQueue(ts.ctx, ts.queueName, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := ts.service.AddMessageToQueue(ts.ctx, ts.testParams)
	assert.NoError(t, err)
}

func TestAddMessageToQueue_Success_WithFiles(t *testing.T) {
	ts := prepareTest(t)
	defer ts.ctrl.Finish()

	testUUID := uuid.New()
	ts.testParams.Files = []dto.SetFile{
		{UUID: testUUID, Title: "file", Ext: "png"},
	}

	ts.mockBroker.EXPECT().EnsureStreamGroup(ts.ctx, ts.queueName, "reader").Return(nil)
	ts.mockStorage.EXPECT().GetSystemById(ts.ctx, ts.testParams.IDSystem).Return(db.System{Name: "TestSystem"}, nil)
	ts.mockBroker.EXPECT().AddMessageToQueue(ts.ctx, ts.queueName, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := ts.service.AddMessageToQueue(ts.ctx, ts.testParams)
	assert.NoError(t, err)
}

func TestAddMessageToQueue_Fail_EnsureStreamGroup(t *testing.T) {
	ts := prepareTest(t)
	defer ts.ctrl.Finish()

	ts.mockBroker.EXPECT().EnsureStreamGroup(ts.ctx, ts.queueName, "reader").Return(errors.New("ошибка Redis"))

	err := ts.service.AddMessageToQueue(ts.ctx, ts.testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка Redis")
}

func TestAddMessageToQueue_Fail_GetSystemById(t *testing.T) {
	ts := prepareTest(t)
	defer ts.ctrl.Finish()

	ts.mockBroker.EXPECT().EnsureStreamGroup(ts.ctx, ts.queueName, "reader").Return(nil)
	ts.mockStorage.EXPECT().GetSystemById(ts.ctx, ts.testParams.IDSystem).Return(db.System{}, errors.New("система не найдена"))

	err := ts.service.AddMessageToQueue(ts.ctx, ts.testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "система не найдена")
}

func TestAddMessageToQueue_Fail_AddMessageToQueue(t *testing.T) {
	ts := prepareTest(t)
	defer ts.ctrl.Finish()

	ts.mockBroker.EXPECT().EnsureStreamGroup(ts.ctx, ts.queueName, "reader").Return(nil)
	ts.mockStorage.EXPECT().GetSystemById(ts.ctx, ts.testParams.IDSystem).Return(db.System{Name: "TestSystem"}, nil)
	ts.mockBroker.EXPECT().AddMessageToQueue(ts.ctx, ts.queueName, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("ошибка добавления в Redis"))

	err := ts.service.AddMessageToQueue(ts.ctx, ts.testParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка добавления в Redis")
}
