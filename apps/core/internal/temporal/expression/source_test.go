package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

func TestResolveSourceSupportsNestedMessageAndStepPaths(t *testing.T) {
	message := []byte(`{"customer":{"name":"Ada"},"type":"vip"}`)
	stepOutputs := map[int32]map[string]any{
		7: {"profile": map[string]any{"tier": "gold"}},
	}

	got, err := ResolveSource("$.message.value.customer.name", message, stepOutputs)
	require.NoError(t, err)
	assert.Equal(t, "Ada", got)

	got, err = ResolveSource("$.steps.7.output.profile.tier", message, stepOutputs)
	require.NoError(t, err)
	assert.Equal(t, "gold", got)
}

func TestResolveSourceExistsModeDoesNotErrorOnMissingField(t *testing.T) {
	got, exists, err := ResolveSourcePresence("$.message.value.missing", []byte(`{"type":"vip"}`), nil)

	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, got)
}

func TestResolveInputUsesSharedResolver(t *testing.T) {
	got, err := ResolveInput([]temporaltypes.MappingEntry{
		{Target: "name", Source: "$.message.value.customer.name"},
		{Target: "tier", Source: "$.steps.7.output.profile.tier"},
	}, []byte(`{"customer":{"name":"Ada"}}`), map[int32]map[string]any{
		7: {"profile": map[string]any{"tier": "gold"}},
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]any{"name": "Ada", "tier": "gold"}, got)
}
