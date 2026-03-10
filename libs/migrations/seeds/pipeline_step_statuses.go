package seeds

import (
	"context"
	"fmt"

	"github.com/zalberix/cactus/apps/core/storage/db"
)

func init() {
	Register("pipeline_step_statuses", SeedPipelineStepStatuses)
}

func SeedPipelineStepStatuses(ctx context.Context, storage *db.Queries) error {
	statuses := []db.CreatePipelineStepStatusParams{
		{Name: "Ожидание", Slug: "pending"},
		{Name: "В обработке", Slug: "processing"},
		{Name: "Завершён", Slug: "completed"},
		{Name: "Ошибка", Slug: "failed"},
		{Name: "Отменён", Slug: "cancelled"},
	}

	for _, s := range statuses {
		_, err := storage.CreatePipelineStepStatus(ctx, s)
		if err != nil {
			return fmt.Errorf("ошибка при создании статуса %q: %w", s.Slug, err)
		}
	}

	return nil
}
