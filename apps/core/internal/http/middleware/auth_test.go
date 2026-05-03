package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/domain/auth"
	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
)

func testAuthService() *auth.Service {
	cfg := config.JWT{
		Secret:     "test-secret-key-32-chars-minimum!",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 720 * time.Hour,
	}
	return auth.NewService(cfg, nil)
}

func setupRouter(svc *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(svc))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestAuth_ValidToken(t *testing.T) {
	svc := testAuthService()
	r := setupRouter(svc)

	resp, err := svc.IssueTokens(42)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+resp.AccessToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_MissingHeader(t *testing.T) {
	svc := testAuthService()
	r := setupRouter(svc)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ExpiredToken(t *testing.T) {
	cfg := config.JWT{
		Secret:     "test-secret-key-32-chars-minimum!",
		AccessTTL:  -1 * time.Second, // уже истёк
		RefreshTTL: 720 * time.Hour,
	}
	svc := auth.NewService(cfg, nil)

	resp, err := svc.IssueTokens(1)
	require.NoError(t, err)

	// Используем svc с нормальным TTL для middleware
	svc2 := testAuthService()
	r := setupRouter(svc2)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+resp.AccessToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_MalformedToken(t *testing.T) {
	svc := testAuthService()
	r := setupRouter(svc)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.jwt")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := testAuthService()
	resp, err := svc.IssueTokens(99)
	require.NoError(t, err)

	r.Use(middleware.Auth(svc))
	r.GET("/test", func(c *gin.Context) {
		claims := middleware.GetClaims(c)
		require.NotNil(t, claims)
		assert.Equal(t, int32(99), claims.UserID)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+resp.AccessToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
