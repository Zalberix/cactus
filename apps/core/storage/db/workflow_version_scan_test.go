package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type workflowVersionScanDB struct{}

func (workflowVersionScanDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (workflowVersionScanDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return &workflowVersionScanRows{}, nil
}

func (workflowVersionScanDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

type workflowVersionScanRows struct {
	seen bool
	err  error
}

func (r *workflowVersionScanRows) Close() {}

func (r *workflowVersionScanRows) Err() error {
	return r.err
}

func (r *workflowVersionScanRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (r *workflowVersionScanRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *workflowVersionScanRows) Next() bool {
	if r.seen {
		return false
	}
	r.seen = true
	return true
}

func (r *workflowVersionScanRows) Scan(dest ...any) error {
	if len(dest) != 17 {
		r.err = fmt.Errorf("scan destination count mismatch: got %d, want 17", len(dest))
		return r.err
	}
	return nil
}

func (r *workflowVersionScanRows) Values() ([]any, error) {
	return nil, nil
}

func (r *workflowVersionScanRows) RawValues() [][]byte {
	return nil
}

func (r *workflowVersionScanRows) Conn() *pgx.Conn {
	return nil
}

func TestListWorkflowVersionsByWorkflowIDScansAllSelectedColumns(t *testing.T) {
	_, err := New(workflowVersionScanDB{}).ListWorkflowVersionsByWorkflowID(context.Background(), 1)
	require.NoError(t, err)
}

func TestListActiveWorkflowVersionsScansAllSelectedColumns(t *testing.T) {
	_, err := New(workflowVersionScanDB{}).ListActiveWorkflowVersions(context.Background(), 1)
	require.NoError(t, err)
}

var (
	_ pgx.Rows = (*workflowVersionScanRows)(nil)
	_ DBTX     = workflowVersionScanDB{}
)
