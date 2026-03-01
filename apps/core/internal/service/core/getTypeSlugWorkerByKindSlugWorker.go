package core

import (
	"context"
)

func (s *Service) GetTypeSlugWorkerByKindSlugWorker(ctx context.Context, slug string) (string, error) {
	return s.storage.GetTypeSlugWorkerByKindSlugWorker(ctx, slug)
}
