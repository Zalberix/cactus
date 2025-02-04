package pipeline_test

import (
	DTO "cactus/internal/DTO"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	pipelinePkg "cactus/internal/pkg/pipeline"
	"cactus/internal/service/pipeline"
	"cactus/internal/service/pipeline/mocks"
	"cactus/internal/storage/db"
)

type testSetupCreatePipeline struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	mockTx      *mocks.MockStorageTx
	service     *pipeline.Service
}

func prepareTestCreatePipeline(t *testing.T) *testSetupCreatePipeline {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockTx := mocks.NewMockStorageTx(ctrl)

	service := pipeline.New(mockStorage, nil)

	return &testSetupCreatePipeline{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		mockTx:      mockTx,
		service:     service,
	}
}

func TestCreatePipeline_Success_SingleStep(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{
			{Step: 1, Name: "Step 1"},
		},
		Message: DTO.Message{Message: db.Message{ID: 321}},
	}

	expectedPipeline := []DTO.Pipeline{
		{IDMessage: 321, Step: 1, Name: "Step 1", Status: string(pipelinePkg.Wait)},
	}

	ts.mockTx.EXPECT().CreatePipelineStep(ts.ctx, gomock.Any()).Return(db.Pipeline(expectedPipeline[0]), nil)

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.NoError(t, err)
	assert.Len(t, result, len(expectedPipeline))
	assert.Equal(t, expectedPipeline[0].IDMessage, result[0].IDMessage)
}

func TestCreatePipeline_Success_MultipleSteps(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{
			{Step: 1, Name: "Step 1"},
			{Step: 2, Name: "Step 2"},
		},
		Message: DTO.Message{Message: db.Message{ID: 321}},
	}

	expectedPipelines := []DTO.Pipeline{
		{IDMessage: 321, Step: 1, Name: "Step 1", Status: string(pipelinePkg.Wait)},
		{IDMessage: 321, Step: 2, Name: "Step 2", Status: string(pipelinePkg.Wait)},
	}

	ts.mockTx.EXPECT().CreatePipelineStep(ts.ctx, gomock.Any()).Return(db.Pipeline(expectedPipelines[0]), nil)
	ts.mockTx.EXPECT().CreatePipelineStep(ts.ctx, gomock.Any()).Return(db.Pipeline(expectedPipelines[1]), nil)

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.NoError(t, err)
	assert.Len(t, result, len(expectedPipelines))
	assert.Equal(t, expectedPipelines[0].IDMessage, result[0].IDMessage)
	assert.Equal(t, expectedPipelines[1].Step, result[1].Step)
}

func TestCreatePipeline_Fail_EmptyPipeline(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{}, // Пустой список шагов
		Message:  DTO.Message{Message: db.Message{ID: 321}},
	}

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.Error(t, err)
	assert.Equal(t, "количество шагов не может быть меньше 1", err.Error())
	assert.Empty(t, result)
}

func TestCreatePipeline_Fail_CreateStepError(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{
			{Step: 1, Name: "Step 1"},
		},
		Message: DTO.Message{Message: db.Message{ID: 321}},
	}

	ts.mockTx.EXPECT().
		CreatePipelineStep(ts.ctx, gomock.Any()).
		Times(1). // Ожидаем один вызов
		Return(db.Pipeline{}, errors.New("ошибка создания шага"))

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания pipeline шага для Step 1")
	assert.Empty(t, result)
}
