package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type TypeWorker = Channel

type CreateTypeWorkerParams struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type KindWorker struct {
	ID           int32                 `json:"id"`
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	ConfigSchema json.RawMessage       `json:"config_schema"`
	Config       pqtype.NullRawMessage `json:"config"`
}

type CreateKindWorkerParams struct {
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	ConfigSchema json.RawMessage       `json:"config_schema"`
	Config       pqtype.NullRawMessage `json:"config"`
}

// Priority for backward compatibility (was priority table, now system.priority in v0.2.0).
type Priority struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	Weight int32  `json:"weight"`
	Slug   string `json:"slug"`
}

// GetPriorityBySystemIDRow for backward compatibility.
type GetPriorityBySystemIDRow struct {
	Weight int32 `json:"weight"`
}

// Token for backward compatibility (was token table, now system.public_token/private_token in v0.2.0).
type Token struct {
	IDSystem     int32  `json:"id_system"`
	IDKindWorker int32  `json:"id_kind_worker"`
	IsActive     bool   `json:"is_active"`
	PublicToken  string `json:"public_token"`
	SecretToken  string `json:"secret_token"`
}

// GetMessagesByParams for backward compatibility.
type GetMessagesByParams struct {
	IDTypeWorker int32 `json:"id_type_worker"`
	IDSystem     int32 `json:"id_system"`
}

// UpdatePipelineStatusAndWorkerByIDParams for backward compatibility (pipeline_step in v0.2.0).
type UpdatePipelineStatusAndWorkerByIDParams struct {
	ID       int32         `json:"id"`
	Status   string        `json:"status"`
	IDWorker sql.NullInt32 `json:"id_worker"`
}

// GetIDPipelineByUUIDMessageAndStepParams for backward compatibility.
type GetIDPipelineByUUIDMessageAndStepParams struct {
	UUID uuid.UUID `json:"uuid"`
	Step int32     `json:"step"`
}

// ─── Wrapper methods on *Queries ────────────────────────────────────────────

// GetTypeWorkers returns all channels (channel replaced type_worker in v0.2.0).
func (q *Queries) GetTypeWorkers(ctx context.Context) ([]TypeWorker, error) {
	return q.GetChannels(ctx)
}

// GetTypeWorkerBySlug returns a channel by slug.
func (q *Queries) GetTypeWorkerBySlug(ctx context.Context, slug string) (TypeWorker, error) {
	return q.GetChannelBySlug(ctx, slug)
}

// CreateTypeWorker creates a channel (type_worker replacement).
func (q *Queries) CreateTypeWorker(ctx context.Context, arg CreateTypeWorkerParams) (TypeWorker, error) {
	return q.CreateChannel(ctx, CreateChannelParams(arg))
}

// GetKindWorkerBySlug returns a KindWorker by slug (stub — config has no slug in v0.2.0).
func (q *Queries) GetKindWorkerBySlug(_ context.Context, _ string) (KindWorker, error) {
	return KindWorker{}, sql.ErrNoRows
}

// CreateKindWorker creates a Config and returns it as KindWorker.
func (q *Queries) CreateKindWorker(ctx context.Context, arg CreateKindWorkerParams) (KindWorker, error) {
	config, err := q.CreateConfig(ctx, CreateConfigParams{
		Name:         arg.Name,
		ConfigSchema: arg.ConfigSchema,
		Config:       arg.Config,
	})
	if err != nil {
		return KindWorker{}, err
	}
	return KindWorker{
		ID:           config.ID,
		Name:         config.Name,
		Slug:         arg.Slug,
		ConfigSchema: config.ConfigSchema,
		Config:       config.Config,
	}, nil
}

// GetKindWokerByID returns a Config as KindWorker by ID.
func (q *Queries) GetKindWokerByID(ctx context.Context, id int32) (KindWorker, error) {
	config, err := q.GetConfigByID(ctx, id)
	if err != nil {
		return KindWorker{}, err
	}
	return KindWorker{
		ID:           config.ID,
		Name:         config.Name,
		ConfigSchema: config.ConfigSchema,
		Config:       config.Config,
	}, nil
}

// GetPriorityBySystemID returns priority weight from system record.
func (q *Queries) GetPriorityBySystemID(ctx context.Context, id int32) (GetPriorityBySystemIDRow, error) {
	sys, err := q.GetSystemByID(ctx, id)
	if err != nil {
		return GetPriorityBySystemIDRow{}, err
	}
	return GetPriorityBySystemIDRow{Weight: sys.Priority}, nil
}

// GetPriorityBySlug returns a stub Priority (priority table removed in v0.2.0).
func (q *Queries) GetPriorityBySlug(_ context.Context, slug string) (Priority, error) {
	return Priority{ID: 1, Name: slug, Weight: 0, Slug: slug}, nil
}

// GetMaxPriorityWeight returns 0 (priority table removed in v0.2.0).
func (q *Queries) GetMaxPriorityWeight(_ context.Context) (int32, error) {
	return 0, nil
}

// GetTokenByPublicToken looks up the system by public token and returns it as Token.
func (q *Queries) GetTokenByPublicToken(ctx context.Context, publicToken string) (Token, error) {
	sys, err := q.GetSystemByPublicToken(ctx, sql.NullString{String: publicToken, Valid: true})
	if err != nil {
		return Token{}, err
	}
	return Token{
		IDSystem:    sys.ID,
		IsActive:    sys.IsActive,
		PublicToken: sys.PublicToken.String,
	}, nil
}

// GetWorkerByUUID is a stub — Worker no longer has a UUID column in v0.2.0.
func (q *Queries) GetWorkerByUUID(_ context.Context, _ uuid.UUID) (Worker, error) {
	return Worker{}, sql.ErrNoRows
}

// GetMessagesBy returns messages filtered by system ID (IDTypeWorker filter dropped in v0.2.0).
func (q *Queries) GetMessagesBy(ctx context.Context, arg GetMessagesByParams) ([]Message, error) {
	return q.GetMessagesBySystemID(ctx, arg.IDSystem)
}

// GetStatusMessageByUUID is a stub.
func (q *Queries) GetStatusMessageByUUID(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

// GetFilePathByUUID is a stub.
func (q *Queries) GetFilePathByUUID(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

// GetTypeSlugWorkerByKindSlugWorker is a stub.
func (q *Queries) GetTypeSlugWorkerByKindSlugWorker(_ context.Context, _ string) (string, error) {
	return "", nil
}

// UpdatePipelineStatusAndWorkerByID maps old Status string to PipelineStepStatusID.
func (q *Queries) UpdatePipelineStatusAndWorkerByID(
	ctx context.Context,
	arg UpdatePipelineStatusAndWorkerByIDParams,
) (PipelineStep, error) {
	statusID := int32(1) // default: wait
	switch arg.Status {
	case "work":
		statusID = 2
	case "done":
		statusID = 3
	case "cancel":
		statusID = 4
	case "error":
		statusID = 5
	}
	return q.UpdatePipelineStepStatusAndWorker(ctx, UpdatePipelineStepStatusAndWorkerParams{
		ID:                   arg.ID,
		PipelineStepStatusID: statusID,
		WorkerID:             arg.IDWorker,
	})
}

// GetIDPipelineByUUIDMessageAndStep finds a PipelineStep.ID by message UUID and step number.
func (q *Queries) GetIDPipelineByUUIDMessageAndStep(
	ctx context.Context,
	arg GetIDPipelineByUUIDMessageAndStepParams,
) (int32, error) {
	msg, err := q.GetMessageByUUID(ctx, arg.UUID)
	if err != nil {
		return 0, err
	}
	pipelines, err := q.GetPipelinesByMessageID(ctx, msg.ID)
	if err != nil || len(pipelines) == 0 {
		return 0, sql.ErrNoRows
	}
	step, err := q.GetPipelineStepByPipelineIDAndStep(ctx, GetPipelineStepByPipelineIDAndStepParams{
		PipelineID: pipelines[0].ID,
		Step:       arg.Step,
	})
	if err != nil {
		return 0, err
	}
	return step.ID, nil
}
