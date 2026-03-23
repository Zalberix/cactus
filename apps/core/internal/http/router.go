package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/core/config"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	if cfg.Env == config.AppEnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginLogger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func NewHTTPServer(cfg *config.Config, r *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
}

func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		slog.Info("HTTP",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
		)
	}
}
