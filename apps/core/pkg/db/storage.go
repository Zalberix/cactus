package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zalberix/cactus/apps/core/config"
)

func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, url)
}

func NewFx(cfg *config.Config) (*pgxpool.Pool, error) {
	return New(context.Background(), cfg.Database.URL)
}
