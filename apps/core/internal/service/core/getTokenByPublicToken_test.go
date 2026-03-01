package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"cactus/apps/core/internal/service/core"
	"cactus/apps/core/internal/service/core/mocks"
	"cactus/apps/core/storage/db"
)

type testSetupGetToken struct {
	ctrl        *gomock.Controller
	ctx         context.Context
	mockStorage *mocks.MockStorage
	service     *core.Service
}

func prepareTestGetToken(t *testing.T) *testSetupGetToken {
	t.Helper()
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(ctrl)

	service := core.New(mockStorage, nil, nil, nil, nil)

	return &testSetupGetToken{
		ctrl:        ctrl,
		ctx:         ctx,
		mockStorage: mockStorage,
		service:     service,
	}
}

func TestGetTokenByPublicToken_Success(t *testing.T) {
	ts := prepareTestGetToken(t)
	defer ts.ctrl.Finish()

	testToken := "valid-token"
	expectedToken := db.Token{PublicToken: testToken}

	ts.mockStorage.EXPECT().GetTokenByPublicToken(ts.ctx, testToken).Return(expectedToken, nil)

	result, err := ts.service.GetTokenByPublicToken(ts.ctx, testToken)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, result)
}

func TestGetTokenByPublicToken_Fail_NotFound(t *testing.T) {
	ts := prepareTestGetToken(t)
	defer ts.ctrl.Finish()

	testToken := "non-existent-token"

	ts.mockStorage.EXPECT().GetTokenByPublicToken(ts.ctx, testToken).Return(db.Token{}, errors.New("токен не найден"))

	result, err := ts.service.GetTokenByPublicToken(ts.ctx, testToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "токен не найден")
	assert.Equal(t, db.Token{}, result)
}

func TestGetTokenByPublicToken_Fail_StorageError(t *testing.T) {
	ts := prepareTestGetToken(t)
	defer ts.ctrl.Finish()

	testToken := "error-token"

	ts.mockStorage.EXPECT().GetTokenByPublicToken(ts.ctx, testToken).Return(db.Token{}, errors.New("ошибка базы данных"))

	result, err := ts.service.GetTokenByPublicToken(ts.ctx, testToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка базы данных")
	assert.Equal(t, db.Token{}, result)
}
