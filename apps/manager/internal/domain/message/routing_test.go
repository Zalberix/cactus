package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProcessRefAcceptsWorkflowAndSchemaVersion(t *testing.T) {
	got, err := parseProcessRef("1.v2")
	require.NoError(t, err)
	require.Equal(t, int32(1), got.WorkflowID)
	require.Equal(t, int32(2), got.InputSchemaVersion)
}

func TestParseProcessRefRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"", "1", "1.2", "abc.v2", "1.v0", "0.v1"} {
		_, err := parseProcessRef(raw)
		require.Error(t, err, raw)
	}
}
