package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type Store struct {
	pool *pgxpool.Pool
	*db.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: db.New(pool),
	}
}

func (s *Store) WithTx(ctx context.Context, fn func(q *db.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(db.New(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}
