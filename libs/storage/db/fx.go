package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type Opts struct {
	fx.In
	Pool *pgxpool.Pool
}

func NewFx(opts Opts) *Queries {
	return New(opts.Pool)
}
