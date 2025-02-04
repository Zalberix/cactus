package core_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	mocks_plugin "cactus/internal/plugin/mocks"
	"cactus/internal/service/core"
	"cactus/internal/service/core/mocks"
	"cactus/internal/storage/db"
)

type testSetupGetMessages struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	mockPlugins *mocks.MockPlugins
	service     *core.Service
	slug        string
	systemID    int
}

func prepareTestGetMessages(t *testing.T) *testSetupGetMessages {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockPlugins := mocks.NewMockPlugins(ctrl)

	service := core.New(mockStorage, nil, nil, mockPlugins, nil)

	return &testSetupGetMessages{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		mockPlugins: mockPlugins,
		service:     service,
		slug:        "test_slug",
		systemID:    1,
	}
}

func TestGetMessages_Success(t *testing.T) {
	ts := prepareTestGetMessages(t)
	defer ts.ctrl.Finish()

	expectedTypeWorker := db.TypeWorker{ID: 1}
	expectedMessagesDB := []db.Message{
		{ID: 1, Value: json.RawMessage(`{"key":"value"}`)},
	}

	mockPlugin := mocks_plugin.NewMockPlugin(ts.ctrl)
	expectedSchema := mockPlugin

	ts.mockStorage.EXPECT().GetTypeWorkerBySlug(ts.ctx, ts.slug).Return(expectedTypeWorker, nil)
	ts.mockStorage.EXPECT().
		GetMessagesBy(ts.ctx, db.GetMessagesByParams{
			IDTypeWorker: expectedTypeWorker.ID,
			IDSystem:     int32(ts.systemID),
		}).
		Return(expectedMessagesDB, nil)
	ts.mockPlugins.EXPECT().Get(ts.slug).Return(expectedSchema, true)

	result, err := ts.service.GetMessages(ts.ctx, ts.slug, ts.systemID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetMessages_Success_EmptyMessages(t *testing.T) {
	ts := prepareTestGetMessages(t)
	defer ts.ctrl.Finish()

	expectedTypeWorker := db.TypeWorker{ID: 1}

	ts.mockStorage.EXPECT().
		GetTypeWorkerBySlug(ts.ctx, ts.slug).
		Return(expectedTypeWorker, nil)

	ts.mockStorage.EXPECT().
		GetMessagesBy(ts.ctx, db.GetMessagesByParams{
			IDTypeWorker: expectedTypeWorker.ID,
			IDSystem:     int32(ts.systemID),
		}).
		Return([]db.Message{}, nil)

	result, err := ts.service.GetMessages(ts.ctx, ts.slug, ts.systemID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestGetMessages_Fail_GetTypeWorkerBySlug(t *testing.T) {
	ts := prepareTestGetMessages(t)
	defer ts.ctrl.Finish()

	ts.mockStorage.EXPECT().
		GetTypeWorkerBySlug(ts.ctx, ts.slug).
		Return(db.TypeWorker{}, errors.New("ошибка получения типа воркера"))

	result, err := ts.service.GetMessages(ts.ctx, ts.slug, ts.systemID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения типа воркера")
	assert.Empty(t, result)
}

func TestGetMessages_Fail_GetMessagesBy(t *testing.T) {
	ts := prepareTestGetMessages(t)
	defer ts.ctrl.Finish()

	expectedTypeWorker := db.TypeWorker{ID: 1}

	ts.mockStorage.EXPECT().GetTypeWorkerBySlug(ts.ctx, ts.slug).Return(expectedTypeWorker, nil)
	ts.mockStorage.EXPECT().
		GetMessagesBy(ts.ctx, db.GetMessagesByParams{
			IDTypeWorker: expectedTypeWorker.ID,
			IDSystem:     int32(ts.systemID),
		}).
		Return(nil, errors.New("ошибка получения сообщений"))

	result, err := ts.service.GetMessages(ts.ctx, ts.slug, ts.systemID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения сообщений")
	assert.Empty(t, result)
}

func TestGetMessages_Fail_PluginNotFound(t *testing.T) {
	ts := prepareTestGetMessages(t)
	defer ts.ctrl.Finish()

	expectedTypeWorker := db.TypeWorker{ID: 1}
	expectedMessagesDB := []db.Message{
		{ID: 1, Value: json.RawMessage(`{"key":"value"}`)},
	}

	ts.mockStorage.EXPECT().GetTypeWorkerBySlug(ts.ctx, ts.slug).Return(expectedTypeWorker, nil)
	ts.mockStorage.EXPECT().
		GetMessagesBy(ts.ctx, db.GetMessagesByParams{
			IDTypeWorker: expectedTypeWorker.ID,
			IDSystem:     int32(ts.systemID),
		}).
		Return(expectedMessagesDB, nil)

	ts.mockPlugins.EXPECT().Get(ts.slug).Return(nil, false)

	result, err := ts.service.GetMessages(ts.ctx, ts.slug, ts.systemID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("схема удалена: %v", ts.slug))
	assert.Empty(t, result)
}
