package db

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
)

type Opts struct {
	fx.In
	DB *sqlx.DB
}

func NewFx(opts Opts) *Queries {
	return New(opts.DB)
}
