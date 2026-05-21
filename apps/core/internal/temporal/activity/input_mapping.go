package activity

import (
	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	"github.com/zalberix/cactus/apps/core/internal/temporal/expression"
)

// ResolveInput forms step input data from its input_mapping.
func ResolveInput(mapping []temporaltypes.MappingEntry, messageValue []byte, stepOutputs map[int32]map[string]any) (map[string]any, error) {
	return expression.ResolveInput(mapping, messageValue, stepOutputs)
}
