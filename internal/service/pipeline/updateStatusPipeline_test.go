package pipeline_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	DTO "cactus/internal/DTO"
	pipelinePkg "cactus/internal/pkg/pipeline"
	"cactus/internal/service/pipeline"
	"cactus/internal/service/pipeline/mocks"
	"cactus/internal/storage/db"
)

type testSetupUpdateStatusPipeline struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	service     *pipeline.Service
}

func prepareTestUpdateStatusPipeline(t *testing.T) *testSetupUpdateStatusPipeline {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)

	service := pipeline.New(mockStorage, nil)

	return &testSetupUpdateStatusPipeline{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		service:     service,
	}
}

func TestUpdateStatusPipeline_Success(t *testing.T) {
	ts := prepareTestUpdateStatusPipeline(t)
	defer ts.ctrl.Finish()

	const uuidMessage = "550e8400-e29b-41d4-a716-446655440000"
	const uuidWorker = "6f9619ff-8b86-d011-b42d-00c04fc964ff"
	step := int32(1)
	status := pipelinePkg.Work

	messageUUID, _ := uuid.Parse(uuidMessage)
	workerUUID, _ := uuid.Parse(uuidWorker)

	mockPipelineID := int32(123)
	mockWorker := db.Worker{ID: 42}
	mockUpdatedPipeline := db.Pipeline{
		ID:       mockPipelineID,
		Status:   string(status),
		IDWorker: sql.NullInt32{Valid: true, Int32: mockWorker.ID},
	}

	ts.mockStorage.EXPECT().
		GetIdPipelineByUUIDMessageAndStep(ts.ctx, db.GetIdPipelineByUUIDMessageAndStepParams{Uuid: messageUUID, Step: step}).
		Times(1).
		Return(mockPipelineID, nil)

	ts.mockStorage.EXPECT().
		GetWorkerByUUID(ts.ctx, workerUUID).
		Times(1).
		Return(mockWorker, nil)

	ts.mockStorage.EXPECT().
		UpdatePipelineStatusAndWorkerByID(ts.ctx, db.UpdatePipelineStatusAndWorkerByIDParams{
			ID:     mockPipelineID,
			Status: string(status),
			IDWorker: sql.NullInt32{
				Valid: true,
				Int32: mockWorker.ID,
			},
		}).
		Times(1).
		Return(mockUpdatedPipeline, nil)

	result, err := ts.service.UpdateStatusPipeline(ts.ctx, uuidMessage, step, status, uuidWorker)

	assert.NoError(t, err)
	assert.Equal(t, DTO.Pipeline(mockUpdatedPipeline), result)
}

func TestUpdateStatusPipeline_Fail_InvalidMessageUUID(t *testing.T) {
	ts := prepareTestUpdateStatusPipeline(t)
	defer ts.ctrl.Finish()

	invalidUUID := "invalid-uuid"
	step := int32(1)
	status := pipelinePkg.Work
	const workerUUID = "6f9619ff-8b86-d011-b42d-00c04fc964ff"

	result, err := ts.service.UpdateStatusPipeline(ts.ctx, invalidUUID, step, status, workerUUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "проверьте UUID сообщения, он неверного формата")
	assert.Empty(t, result)
}

func TestUpdateStatusPipeline_Fail_StepNotFound(t *testing.T) {
	ts := prepareTestUpdateStatusPipeline(t)
	defer ts.ctrl.Finish()

	uuidMessage := "550e8400-e29b-41d4-a716-446655440000"
	uuidWorker := "6f9619ff-8b86-d011-b42d-00c04fc964ff"
	step := int32(99)
	status := pipelinePkg.Work

	messageUUID, _ := uuid.Parse(uuidMessage)

	ts.mockStorage.EXPECT().
		GetIdPipelineByUUIDMessageAndStep(ts.ctx, db.GetIdPipelineByUUIDMessageAndStepParams{Uuid: messageUUID, Step: step}).
		Times(1).
		Return(int32(0), errors.New("шаг не найден"))

	result, err := ts.service.UpdateStatusPipeline(ts.ctx, uuidMessage, step, status, uuidWorker)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "шаг пайплайна не найден по uuid сообщения и номеру шага")
	assert.Empty(t, result)
}

func TestUpdateStatusPipeline_Fail_WorkerNotFound(t *testing.T) {
	ts := prepareTestUpdateStatusPipeline(t)
	defer ts.ctrl.Finish()

	uuidMessage := "550e8400-e29b-41d4-a716-446655440000"
	uuidWorker := "6f9619ff-8b86-d011-b42d-00c04fc964ff"
	step := int32(2)
	status := pipelinePkg.Error

	messageUUID, _ := uuid.Parse(uuidMessage)
	workerUUID, _ := uuid.Parse(uuidWorker)

	mockPipelineID := int32(124)

	ts.mockStorage.EXPECT().
		GetIdPipelineByUUIDMessageAndStep(ts.ctx, db.GetIdPipelineByUUIDMessageAndStepParams{Uuid: messageUUID, Step: step}).
		Times(1).
		Return(mockPipelineID, nil)

	ts.mockStorage.EXPECT().
		GetWorkerByUUID(ts.ctx, workerUUID).
		Times(1).
		Return(db.Worker{}, errors.New("воркер не найден"))

	result, err := ts.service.UpdateStatusPipeline(ts.ctx, uuidMessage, step, status, uuidWorker)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "воркер с uuid 6f9619ff-8b86-d011-b42d-00c04fc964ff не зарегистрирован")
	assert.Empty(t, result)
}
