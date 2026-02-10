package seeds

import (
	"context"
	"fmt"

	"cactus/internal/storage/db"
)

func init() {
	Register("all", SeedAll)
}

func SeedAll(ctx context.Context, storage *db.Queries) error {
	ordered := []struct {
		name string
		fn   SeedFunc
	}{
		{"channels", SeedChannels},
		{"pipeline_step_statuses", SeedPipelineStepStatuses},
		{"users", SeedUsers},
		{"configs", SeedConfigs},
		{"systems", SeedSystems},
	}

	for _, s := range ordered {
		if err := s.fn(ctx, storage); err != nil {
			return fmt.Errorf("сид %q: %w", s.name, err)
		}
	}

	return nil
}
