package activity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

func TestResolveInput(t *testing.T) {
	tests := []struct {
		name        string
		mapping     []temporaltypes.MappingEntry
		want        map[string]any
		wantErr     bool
		messageJSON []byte
	}{
		{
			name:        "message value resolves",
			mapping:     []temporaltypes.MappingEntry{{Target: "to", Source: "$.message.value.to"}},
			messageJSON: []byte(`{"to":"person@example.com"}`),
			want:        map[string]any{"to": "person@example.com"},
		},
		{
			name:        "step output resolves",
			mapping:     []temporaltypes.MappingEntry{{Target: "body", Source: "$.steps.1.output.body"}},
			messageJSON: []byte(`{}`),
			want:        map[string]any{"body": "Rendered"},
		},
		{
			name:        "static source resolves",
			mapping:     []temporaltypes.MappingEntry{{Target: "subject", Source: "Hello"}},
			messageJSON: []byte(`{}`),
			want:        map[string]any{"subject": "Hello"},
		},
		{
			name:        "duplicate target errors",
			mapping:     []temporaltypes.MappingEntry{{Target: "to", Source: "$.message.value.to"}, {Target: "to", Source: "Hello"}},
			messageJSON: []byte(`{"to":"person@example.com"}`),
			wantErr:     true,
		},
		{
			name:        "nested message source errors",
			mapping:     []temporaltypes.MappingEntry{{Target: "to", Source: "$.message.value.customer.name"}},
			messageJSON: []byte(`{"customer":{"name":"Ada"}}`),
			wantErr:     true,
		},
		{
			name:        "nested step output source errors",
			mapping:     []temporaltypes.MappingEntry{{Target: "name", Source: "$.steps.1.output.customer.name"}},
			messageJSON: []byte(`{}`),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveInput(tt.mapping, tt.messageJSON, map[int32]map[string]any{
				1: {"body": "Rendered", "customer": map[string]any{"name": "Ada"}},
			})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveInputEmptyMappingReturnsNil(t *testing.T) {
	got, err := ResolveInput(nil, []byte(`not-json`), nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}
