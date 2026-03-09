package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	DTO "github.com/zalberix/cactus/apps/core/internal/DTO"
	pipelinePkg "github.com/zalberix/cactus/libs/pipeline"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline/mocks"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

type testSetupCreatePipeline struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	mockTx      *mocks.MockStorageTx
	service     *pipeline.Service
}

func prepareTestCreatePipeline(t *testing.T) *testSetupCreatePipeline {
	t.Helper()
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
		Message:   DTO.Message{Message: db.Message{ID: 321}},
		ChannelID: 1,
	}

	mockParentPipeline := db.Pipeline{ID: 100, MessageID: 321}
	expectedStep := db.PipelineStep{ID: 1, PipelineID: 100, Step: 1, PipelineStepStatusID: 1}

	ts.mockTx.EXPECT().CreatePipeline(ts.ctx, gomock.Any()).Return(mockParentPipeline, nil)
	ts.mockTx.EXPECT().CreatePipelineStep(ts.ctx, gomock.Any()).Return(expectedStep, nil)

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedStep.PipelineID, result[0].PipelineID)
	assert.Equal(t, expectedStep.Step, result[0].Step)
}

func TestCreatePipeline_Success_MultipleSteps(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{
			{Step: 1, Name: "Step 1"},
			{Step: 2, Name: "Step 2"},
		},
		Message:   DTO.Message{Message: db.Message{ID: 321}},
		ChannelID: 1,
	}

	mockParentPipeline := db.Pipeline{ID: 100, MessageID: 321}
	step1 := db.PipelineStep{ID: 1, PipelineID: 100, Step: 1, PipelineStepStatusID: 1}
	step2 := db.PipelineStep{ID: 2, PipelineID: 100, Step: 2, PipelineStepStatusID: 1}

	ts.mockTx.EXPECT().CreatePipeline(ts.ctx, gomock.Any()).Return(mockParentPipeline, nil)
	ts.mockTx.EXPECT().CreatePipelineStep(ts.ctx, gomock.Any()).Return(step1, nil)
	ts.mockTx.EXPECT().CreatePipelineStep(ts.ctx, gomock.Any()).Return(step2, nil)

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, step1.PipelineID, result[0].PipelineID)
	assert.Equal(t, step2.Step, result[1].Step)
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

func TestCreatePipeline_Fail_CreatePipelineError(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{
			{Step: 1, Name: "Step 1"},
		},
		Message:   DTO.Message{Message: db.Message{ID: 321}},
		ChannelID: 1,
	}

	ts.mockTx.EXPECT().
		CreatePipeline(ts.ctx, gomock.Any()).
		Return(db.Pipeline{}, errors.New("ошибка создания pipeline"))

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания pipeline")
	assert.Empty(t, result)
}

func TestCreatePipeline_Fail_CreateStepError(t *testing.T) {
	ts := prepareTestCreatePipeline(t)
	defer ts.ctrl.Finish()

	arg := pipeline.CreatePipelineParams{
		Pipeline: []pipelinePkg.Step{
			{Step: 1, Name: "Step 1"},
		},
		Message:   DTO.Message{Message: db.Message{ID: 321}},
		ChannelID: 1,
	}

	mockParentPipeline := db.Pipeline{ID: 100, MessageID: 321}

	ts.mockTx.EXPECT().
		CreatePipeline(ts.ctx, gomock.Any()).
		Return(mockParentPipeline, nil)

	ts.mockTx.EXPECT().
		CreatePipelineStep(ts.ctx, gomock.Any()).
		Times(1).
		Return(db.PipelineStep{}, errors.New("ошибка создания шага"))

	result, err := ts.service.CreatePipelineTX(ts.ctx, ts.mockTx, arg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания pipeline шага для Step 1")
	assert.Empty(t, result)
}

