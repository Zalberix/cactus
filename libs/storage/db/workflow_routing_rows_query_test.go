package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListWorkflowRoutingVersionRowsQueryIsInputSchemaCentric(t *testing.T) {
	query := listWorkflowRoutingVersionRows

	require.Contains(t, query, `FROM "workflow_input_schema" wis`)
	require.Contains(t, query, `AS supported_versions`)
	require.NotContains(t, query, `JOIN native_schema`)
	require.NotContains(t, query, `AS supported_schemas`)
}

func TestListWorkflowRoutingVersionRowsQueryIsServerPaginated(t *testing.T) {
	query := listWorkflowRoutingVersionRows

	require.Contains(t, query, `LIMIT $2`)
	require.Contains(t, query, `OFFSET $3`)
	require.Contains(t, countWorkflowRoutingVersionRows, `COUNT(*)`)
	require.Contains(t, countWorkflowRoutingVersionRows, `wis.workflow_id = $1`)
	require.NotContains(t, countWorkflowRoutingVersionRows, `GROUP BY`)
}
