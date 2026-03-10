package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/zalberix/cactus/apps/core/config"
)

func New(
	ctx context.Context,
	url string,
) (*sqlx.DB, error) {
	return sqlx.ConnectContext(ctx, "postgres", url)
}

func NewFx(cfg *config.Config) (*sqlx.DB, error) {
	return New(
		context.Background(),
		cfg.Database.URL,
	)
}
