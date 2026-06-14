package message

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/libs/storage/db"
)

type inputSchemaReadStore interface {
	GetWorkflowInputSchemaByID(ctx context.Context, id int32) (db.WorkflowInputSchema, error)
	GetWorkflowInputSchemaByCode(ctx context.Context, arg db.GetWorkflowInputSchemaByCodeParams) (db.WorkflowInputSchema, error)
	GetWorkflowInputSchemaByVersionNumber(ctx context.Context, arg db.GetWorkflowInputSchemaByVersionNumberParams) (db.WorkflowInputSchema, error)
	GetDefaultWorkflowInputSchema(ctx context.Context, workflowID int32) (db.WorkflowInputSchema, error)
}

type messageIdempotencyStore interface {
	GetMessageByIdempotencyKey(ctx context.Context, arg db.GetMessageByIdempotencyKeyParams) (db.Message, error)
	GetLatestWorkflowRunByMessageID(ctx context.Context, messageID int32) (db.WorkflowRun, error)
}

type compatibilityLookupStore interface {
	GetActiveWorkflowVersionInputSchemaCompatibilityByPair(ctx context.Context, arg db.GetActiveWorkflowVersionInputSchemaCompatibilityByPairParams) (db.WorkflowVersionInputSchemaCompatibility, error)
}

type resolvedInputSchema struct {
	ID            pgtype.Int4
	Code          string
	VersionNumber pgtype.Int4
	SchemaJSON    []byte
}

type processRef struct {
	WorkflowID         int32
	InputSchemaVersion int32
}

func parseProcessRef(raw string) (processRef, error) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return processRef{}, fmt.Errorf("PROCESS_INVALID: expected workflowId.vSchemaVersion")
	}

	workflowID64, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil || workflowID64 <= 0 {
		return processRef{}, fmt.Errorf("PROCESS_INVALID: workflow id must be positive")
	}

	versionPart := strings.ToLower(strings.TrimSpace(parts[1]))
	if !strings.HasPrefix(versionPart, "v") {
		return processRef{}, fmt.Errorf("PROCESS_INVALID: schema version must be formatted as vN")
	}
	versionText := strings.TrimPrefix(versionPart, "v")
	schemaVersion64, err := strconv.ParseInt(versionText, 10, 32)
	if err != nil || schemaVersion64 <= 0 {
		return processRef{}, fmt.Errorf("PROCESS_INVALID: schema version must be formatted as vN")
	}

	return processRef{
		WorkflowID:         int32(workflowID64),
		InputSchemaVersion: int32(schemaVersion64),
	}, nil
}

func (s *Service) resolveInputSchema(ctx context.Context, req SendMessageRequest, wf db.Workflow) (resolvedInputSchema, error) {
	reader, ok := s.store.(inputSchemaReadStore)
	if !ok {
		return resolvedInputSchema{}, fmt.Errorf("INPUT_SCHEMA_RESOLUTION_UNSUPPORTED")
	}

	if strings.TrimSpace(req.Process) != "" {
		ref, err := parseProcessRef(req.Process)
		if err != nil {
			return resolvedInputSchema{}, err
		}
		if req.WorkflowID != 0 && req.WorkflowID != ref.WorkflowID {
			return resolvedInputSchema{}, fmt.Errorf("PROCESS_WORKFLOW_MISMATCH")
		}
		req.WorkflowID = ref.WorkflowID
		schema, err := reader.GetWorkflowInputSchemaByVersionNumber(ctx, db.GetWorkflowInputSchemaByVersionNumberParams{
			WorkflowID:    ref.WorkflowID,
			VersionNumber: ref.InputSchemaVersion,
		})
		if err != nil {
			return resolvedInputSchema{}, fmt.Errorf("resolve process input schema: %w", err)
		}
		if schema.WorkflowID != wf.ID {
			return resolvedInputSchema{}, fmt.Errorf("INPUT_SCHEMA_WORKFLOW_MISMATCH")
		}
		if err := ensureInputSchemaAcceptsMessages(schema, true); err != nil {
			return resolvedInputSchema{}, err
		}
		return resolvedInputSchema{
			ID:            pgtype.Int4{Int32: schema.ID, Valid: true},
			Code:          schema.Code,
			VersionNumber: pgtype.Int4{Int32: schema.VersionNumber, Valid: true},
			SchemaJSON:    schema.SchemaJson,
		}, nil
	}

	if req.InputSchemaID != nil && strings.TrimSpace(req.InputSchemaCode) != "" {
		return resolvedInputSchema{}, fmt.Errorf("INPUT_SCHEMA_AMBIGUOUS: use input_schema_id or input_schema_code, not both")
	}

	var schema db.WorkflowInputSchema
	var err error
	explicit := false
	switch {
	case req.InputSchemaID != nil:
		explicit = true
		schema, err = reader.GetWorkflowInputSchemaByID(ctx, *req.InputSchemaID)
	case strings.TrimSpace(req.InputSchemaCode) != "":
		explicit = true
		schema, err = reader.GetWorkflowInputSchemaByCode(ctx, db.GetWorkflowInputSchemaByCodeParams{
			WorkflowID: req.WorkflowID,
			Code:       strings.TrimSpace(req.InputSchemaCode),
		})
	default:
		schema, err = reader.GetDefaultWorkflowInputSchema(ctx, req.WorkflowID)
	}
	if err != nil {
		return resolvedInputSchema{}, fmt.Errorf("resolve input schema: %w", err)
	}
	if schema.WorkflowID != wf.ID {
		return resolvedInputSchema{}, fmt.Errorf("INPUT_SCHEMA_WORKFLOW_MISMATCH")
	}
	if err := ensureInputSchemaAcceptsMessages(schema, explicit); err != nil {
		return resolvedInputSchema{}, err
	}
	return resolvedInputSchema{
		ID:            pgtype.Int4{Int32: schema.ID, Valid: true},
		Code:          schema.Code,
		VersionNumber: pgtype.Int4{Int32: schema.VersionNumber, Valid: true},
		SchemaJSON:    schema.SchemaJson,
	}, nil
}

func ensureInputSchemaAcceptsMessages(schema db.WorkflowInputSchema, explicit bool) error {
	switch schema.Status {
	case "active":
		return nil
	case "deprecated":
		if explicit {
			return nil
		}
		return fmt.Errorf("INPUT_SCHEMA_DEPRECATED: deprecated schema must be requested explicitly")
	case "draft":
		return fmt.Errorf("INPUT_SCHEMA_DRAFT: draft schema cannot receive messages")
	case "archived":
		return fmt.Errorf("INPUT_SCHEMA_ARCHIVED: archived schema cannot receive messages")
	default:
		return fmt.Errorf("INPUT_SCHEMA_STATUS_INVALID: %s", schema.Status)
	}
}

func (s *Service) findExistingIdempotentMessage(ctx context.Context, req SendMessageRequest, valueJSON []byte) (*SendMessageResponse, bool, error) {
	key := strings.TrimSpace(req.IdempotencyKey)
	if key == "" {
		return nil, false, nil
	}
	store, ok := s.store.(messageIdempotencyStore)
	if !ok {
		return nil, false, nil
	}

	msg, err := store.GetMessageByIdempotencyKey(ctx, db.GetMessageByIdempotencyKeyParams{
		WorkflowID:     req.WorkflowID,
		IdempotencyKey: pgtype.Text{String: key, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("lookup idempotent message: %w", err)
	}
	if !bytes.Equal(bytes.TrimSpace(msg.Value), bytes.TrimSpace(valueJSON)) {
		return nil, false, fmt.Errorf("IDEMPOTENCY_CONFLICT: existing message payload differs")
	}

	resp := &SendMessageResponse{
		MessageID:             msg.ID,
		WorkflowInputSchemaID: pgInt4Ptr(msg.WorkflowInputSchemaID),
		Status:                msg.Status,
	}
	if run, err := store.GetLatestWorkflowRunByMessageID(ctx, msg.ID); err == nil {
		resp.WorkflowRunID = run.ID
		resp.WorkflowVersionID = int32Ptr(run.WorkflowVersionID)
		resp.SelectionReason = run.SelectionReason
		resp.WorkflowExperimentID = pgInt4Ptr(run.WorkflowExperimentID)
		resp.WorkflowExperimentVariantID = pgInt4Ptr(run.WorkflowExperimentVariantID)
		resp.Status = run.Status
	}
	return resp, true, nil
}

func (s *Service) lookupCompatibilityID(ctx context.Context, versionID int32, schemaID pgtype.Int4) pgtype.Int4 {
	if !schemaID.Valid {
		return pgtype.Int4{}
	}
	store, ok := s.store.(compatibilityLookupStore)
	if !ok {
		return pgtype.Int4{}
	}
	compatibility, err := store.GetActiveWorkflowVersionInputSchemaCompatibilityByPair(ctx, db.GetActiveWorkflowVersionInputSchemaCompatibilityByPairParams{
		WorkflowVersionID:     versionID,
		WorkflowInputSchemaID: schemaID.Int32,
	})
	if err != nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: compatibility.ID, Valid: true}
}

func int32Ptr(value int32) *int32 {
	return &value
}
