package store

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/storage/db"
)

type Opts struct {
	fx.In
	DB      *sqlx.DB
	Queries *db.Queries
}

func NewFx(opts Opts) *Store {
	return New(opts.DB, opts.Queries)
}

type Store struct {
	connect *sqlx.DB
	tx      *sql.Tx
	*db.Queries
}

func New(connect *sqlx.DB, queries *db.Queries) *Store {
	return &Store{connect: connect, Queries: queries}
}

func (q *Store) SetContext(ctx context.Context, db interface{}) error {
	tx, err := q.connect.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(db)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("SetContext: db должен быть указателем на интерфейс Storage")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Interface {
		return fmt.Errorf("SetContext: db должен быть интерфейсом Storage")
	}

	realVal := elem.Elem()
	if realVal.Kind() != reflect.Ptr || realVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("SetContext: реализация Storage должна быть структурой")
	}

	originalStore := (*Store)(unsafe.Pointer(realVal.Pointer()))
	newStore := *originalStore

	newStore.tx = tx
	newStore.Queries = q.Queries.WithTx(tx)

	elem.Set(reflect.ValueOf(&newStore))

	return nil
}

func (q *Store) Rollback() error {
	return q.tx.Rollback()
}

func (q *Store) Commit() error {
	return q.tx.Commit()
}
