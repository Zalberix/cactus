package dto

import (
	"github.com/zalberix/cactus/apps/core/storage/db"
)

// Pipeline wraps db.PipelineStep (renamed from pipeline table in schema v0.2.0).
type Pipeline db.PipelineStep
