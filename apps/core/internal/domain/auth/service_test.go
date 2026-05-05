package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/domain/auth"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// mockStorage реализует auth.Storage для тестов без БД.
type mockStorage struct {
	user db.User
	err  error
}

func (m *mockStorage) GetUserByEmail(_ context.Context, _ string) (db.User, error) {
	return m.user, m.err
}

func (m *mockStorage) GetUserByID(_ context.Context, _ int32) (db.User, error) {
	return m.user, m.err
}

func testJWTCfg() config.JWT {
	return config.JWT{
		Secret:     "test-secret-key-32-chars-minimum!",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 720 * time.Hour,
	}
}

func TestIssueTokens(t *testing.T) {
	svc := auth.NewService(testJWTCfg(), &mockStorage{})

	resp, err := svc.IssueTokens(42)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Greater(t, resp.ExpiresIn, int64(0))

	// access token должен содержать правильный user_id
	claims, err := svc.ParseToken(resp.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int32(42), claims.UserID)
}

func TestParseToken_Valid(t *testing.T) {
	svc := auth.NewService(testJWTCfg(), &mockStorage{})

	resp, err := svc.IssueTokens(7)
	require.NoError(t, err)

	claims, err := svc.ParseToken(resp.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int32(7), claims.UserID)
}

func TestParseToken_Expired(t *testing.T) {
	cfg := config.JWT{
		Secret:     "test-secret-key-32-chars-minimum!",
		AccessTTL:  -1 * time.Second, // уже истёк
		RefreshTTL: 720 * time.Hour,
	}
	svc := auth.NewService(cfg, &mockStorage{})

	resp, err := svc.IssueTokens(1)
	require.NoError(t, err)

	_, err = svc.ParseToken(resp.AccessToken)
	assert.Error(t, err)
}

func TestParseToken_WrongKey(t *testing.T) {
	svc1 := auth.NewService(testJWTCfg(), &mockStorage{})
	cfg2 := config.JWT{
		Secret:     "other-secret-key-32-chars-minimum",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 720 * time.Hour,
	}
	svc2 := auth.NewService(cfg2, &mockStorage{})

	resp, err := svc1.IssueTokens(1)
	require.NoError(t, err)

	_, err = svc2.ParseToken(resp.AccessToken)
	assert.Error(t, err)
}

func TestHashPassword_CheckPassword(t *testing.T) {
	svc := auth.NewServiceWithPasswordCost(testJWTCfg(), &mockStorage{}, bcrypt.MinCost)

	hashed, err := svc.HashPassword("mypassword123")
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)

	// правильный пароль — ошибки нет
	err = svc.CheckPassword(hashed, "mypassword123")
	assert.NoError(t, err)

	// неправильный пароль — ошибка
	err = svc.CheckPassword(hashed, "wrongpassword")
	assert.Error(t, err)
}

func TestRefreshToken(t *testing.T) {
	svc := auth.NewService(testJWTCfg(), &mockStorage{})

	// Выдаём первую пару токенов
	resp1, err := svc.IssueTokens(5)
	require.NoError(t, err)

	// Обновляем через refresh token
	resp2, err := svc.Refresh(context.Background(), resp1.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, resp2.AccessToken)
	assert.NotEmpty(t, resp2.RefreshToken)

	// Новый access token должен содержать тот же user_id
	claims, err := svc.ParseToken(resp2.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int32(5), claims.UserID)
}
