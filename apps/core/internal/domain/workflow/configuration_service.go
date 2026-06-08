package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalberix/cactus/apps/core/internal/http/response"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

var (
	ErrWorkflowConfigurationUnsupported = errors.New("workflow configuration storage is unsupported")
	ErrWorkflowConfigurationInvalid     = errors.New("workflow configuration is invalid")
	ErrWorkflowConfigurationInUse       = errors.New("workflow configuration is already referenced")
	ErrWorkflowConfigurationNotFound    = errors.New("workflow configuration not found")
)

type workflowInputSchemaQueries interface {
	ClearDefaultWorkflowInputSchema(ctx context.Context, arg db.ClearDefaultWorkflowInputSchemaParams) error
	CreateWorkflowInputSchema(ctx context.Context, arg db.CreateWorkflowInputSchemaParams) (db.WorkflowInputSchema, error)
	GetDefaultWorkflowInputSchema(ctx context.Context, workflowID int32) (db.WorkflowInputSchema, error)
	GetNextWorkflowInputSchemaVersionNumber(ctx context.Context, workflowID int32) (int32, error)
	GetWorkflowInputSchemaByCode(ctx context.Context, arg db.GetWorkflowInputSchemaByCodeParams) (db.WorkflowInputSchema, error)
	GetWorkflowInputSchemaByID(ctx context.Context, id int32) (db.WorkflowInputSchema, error)
	HasInputSchemaUsage(ctx context.Context, workflowInputSchemaID pgtype.Int4) (bool, error)
	ListWorkflowInputSchemasByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowInputSchema, error)
	SetDefaultWorkflowInputSchema(ctx context.Context, arg db.SetDefaultWorkflowInputSchemaParams) (db.WorkflowInputSchema, error)
	SoftDeleteWorkflowInputSchema(ctx context.Context, id int32) error
	UpdateWorkflowInputSchemaRecord(ctx context.Context, arg db.UpdateWorkflowInputSchemaRecordParams) (db.WorkflowInputSchema, error)
	UpdateWorkflowInputSchemaStatus(ctx context.Context, arg db.UpdateWorkflowInputSchemaStatusParams) (db.WorkflowInputSchema, error)
}

type workflowInputMapperQueries interface {
	CreateWorkflowInputMapper(ctx context.Context, arg db.CreateWorkflowInputMapperParams) (db.WorkflowInputMapper, error)
	GetWorkflowInputMapperByID(ctx context.Context, id int32) (db.WorkflowInputMapper, error)
	ListWorkflowInputMappersByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowInputMapper, error)
	SoftDeleteWorkflowInputMapper(ctx context.Context, id int32) error
	UpdateWorkflowInputMapper(ctx context.Context, arg db.UpdateWorkflowInputMapperParams) (db.WorkflowInputMapper, error)
	UpdateWorkflowInputMapperActive(ctx context.Context, arg db.UpdateWorkflowInputMapperActiveParams) (db.WorkflowInputMapper, error)
}

type workflowCompatibilityQueries interface {
	CreateWorkflowVersionInputSchemaCompatibility(ctx context.Context, arg db.CreateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error)
	DeactivateWorkflowVersionInputSchemaCompatibility(ctx context.Context, arg db.DeactivateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error)
	GetWorkflowVersionInputSchemaCompatibilityByID(ctx context.Context, id int32) (db.WorkflowVersionInputSchemaCompatibility, error)
	HasCompatibilityHistoricalRun(ctx context.Context, id int32) (bool, error)
	ListCompatibilitiesByInputSchemaID(ctx context.Context, workflowInputSchemaID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error)
	ListCompatibilitiesByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error)
	ListCompatibilitiesByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error)
	SoftDeleteWorkflowVersionInputSchemaCompatibility(ctx context.Context, id int32) error
	UpdateCompatibilityDefaultRoute(ctx context.Context, arg db.UpdateCompatibilityDefaultRouteParams) (db.WorkflowVersionInputSchemaCompatibility, error)
	UpdateWorkflowVersionInputSchemaCompatibility(ctx context.Context, arg db.UpdateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error)
}

type workflowInputSchemaUsageQueries interface {
	GetInputSchemaUsageSummary(ctx context.Context, workflowInputSchemaID pgtype.Int4) (db.GetInputSchemaUsageSummaryRow, error)
	GetWorkflowInputSchemaByID(ctx context.Context, id int32) (db.WorkflowInputSchema, error)
}

type workflowInputSchemaSearchQueries interface {
	SearchWorkflowInputSchemas(ctx context.Context, arg db.SearchWorkflowInputSchemasParams) ([]db.WorkflowInputSchema, error)
}

type workflowRoutingVersionRowQueries interface {
	ListWorkflowRoutingVersionRows(ctx context.Context, workflowID int32) ([]db.ListWorkflowRoutingVersionRowsRow, error)
	ListActiveWorkflowTestsByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowExperiment, error)
}

type workflowInputSchemaCompatibilityRowQueries interface {
	ListInputSchemaCompatibilityRows(ctx context.Context, workflowInputSchemaID int32) ([]db.ListInputSchemaCompatibilityRowsRow, error)
}

type workflowCompatibilityTargetQueries interface {
	GetWorkflowVersionByNativeInputSchemaID(ctx context.Context, workflowInputSchemaID int32) (db.WorkflowVersion, error)
}

type inputSchemaMutation struct {
	Code          string
	VersionNumber int32
	SchemaJSON    json.RawMessage
	Status        string
	IsDefault     *bool
}

type inputMapperMutation struct {
	MapperType string
	Rules      json.RawMessage
	IsActive   *bool
}

type compatibilityMutation struct {
	WorkflowInputSchemaID int32
	CompatibilityType     string
	WorkflowInputMapperID *int32
	DefaultValues         json.RawMessage
	IsActive              *bool
	IsDefaultRoute        *bool
}

func (s *Service) inputSchemaQueries() (workflowInputSchemaQueries, error) {
	q, ok := s.store.(workflowInputSchemaQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	return q, nil
}

func (s *Service) inputMapperQueries() (workflowInputMapperQueries, error) {
	q, ok := s.store.(workflowInputMapperQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	return q, nil
}

func (s *Service) compatibilityQueries() (workflowCompatibilityQueries, error) {
	q, ok := s.store.(workflowCompatibilityQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	return q, nil
}

func (s *Service) ListWorkflowInputSchemas(ctx context.Context, workflowID int32) ([]db.WorkflowInputSchema, error) {
	q, err := s.inputSchemaQueries()
	if err != nil {
		return nil, err
	}
	return q.ListWorkflowInputSchemasByWorkflowID(ctx, workflowID)
}

func (s *Service) GetWorkflowInputSchemaRecord(ctx context.Context, inputSchemaID int32) (db.WorkflowInputSchema, error) {
	q, err := s.inputSchemaQueries()
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	schema, err := q.GetWorkflowInputSchemaByID(ctx, inputSchemaID)
	if err != nil {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	return schema, nil
}

func (s *Service) SearchWorkflowInputSchemas(ctx context.Context, workflowID int32, query string, excludeInputSchemaID int32) ([]db.WorkflowInputSchema, error) {
	q, ok := s.store.(workflowInputSchemaSearchQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	return q.SearchWorkflowInputSchemas(ctx, db.SearchWorkflowInputSchemasParams{
		WorkflowID: workflowID,
		Column2:    excludeInputSchemaID,
		Column3:    strings.TrimSpace(query),
	})
}

func (s *Service) GetWorkflowInputSchemaUsage(ctx context.Context, inputSchemaID int32) (InputSchemaUsageResponse, error) {
	q, ok := s.store.(workflowInputSchemaUsageQueries)
	if !ok {
		return InputSchemaUsageResponse{}, ErrWorkflowConfigurationUnsupported
	}
	if _, err := q.GetWorkflowInputSchemaByID(ctx, inputSchemaID); err != nil {
		return InputSchemaUsageResponse{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	row, err := q.GetInputSchemaUsageSummary(ctx, pgInt4(inputSchemaID))
	if err != nil {
		return InputSchemaUsageResponse{}, err
	}
	usedByMapper := row.UsedByMapper.Valid && row.UsedByMapper.Bool
	usage := InputSchemaUsageResponse{
		UsedByMessage:    row.UsedByMessage,
		UsedByMapper:     usedByMapper,
		UsedByExperiment: row.UsedByExperiment,
	}
	if usage.UsedByMessage {
		usage.Reasons = append(usage.Reasons, "message")
	}
	if usage.UsedByMapper {
		usage.Reasons = append(usage.Reasons, "mapper")
	}
	if usage.UsedByExperiment {
		usage.Reasons = append(usage.Reasons, "experiment")
	}
	usage.IsReadOnly = len(usage.Reasons) > 0
	return usage, nil
}

func (s *Service) ListWorkflowRoutingVersionRows(ctx context.Context, workflowID int32) ([]RoutingVersionRowResponse, error) {
	q, ok := s.store.(workflowRoutingVersionRowQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	rows, err := q.ListWorkflowRoutingVersionRows(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	result := make([]RoutingVersionRowResponse, 0, len(rows))
	for _, row := range rows {
		supported, err := decodeRoutingSupportedSchemas(row.SupportedSchemas)
		if err != nil {
			return nil, fmt.Errorf("%w: decode supported schemas: %v", ErrWorkflowConfigurationInvalid, err)
		}
		tests, err := q.ListActiveWorkflowTestsByVersionID(ctx, row.WorkflowVersionID)
		if err != nil {
			return nil, err
		}
		activeTests := make([]RoutingActiveTestResponse, 0, len(tests))
		for _, test := range tests {
			activeTests = append(activeTests, RoutingActiveTestResponse{
				ID:             test.ID,
				Name:           test.Name,
				ExperimentType: test.ExperimentType,
				Status:         test.Status,
			})
		}
		name := row.WorkflowVersionName.String
		if !row.WorkflowVersionName.Valid || strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("Version %d", row.WorkflowVersionNumber)
		}
		result = append(result, RoutingVersionRowResponse{
			WorkflowVersionID:        row.WorkflowVersionID,
			WorkflowID:               row.WorkflowID,
			WorkflowVersionName:      name,
			WorkflowVersionNumber:    row.WorkflowVersionNumber,
			IsValid:                  row.IsValid,
			IsActive:                 row.IsActive,
			NativeInputSchemaID:      row.NativeSchemaID,
			NativeInputSchemaCode:    row.NativeSchemaCode,
			NativeInputSchemaVersion: row.NativeSchemaVersionNumber,
			SupportedSchemas:         supported,
			ActiveTests:              activeTests,
		})
	}
	return result, nil
}

func (s *Service) ListInputSchemaCompatibilityRows(ctx context.Context, inputSchemaID int32) ([]InputSchemaCompatibilityRowResponse, error) {
	q, ok := s.store.(workflowInputSchemaCompatibilityRowQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	rows, err := q.ListInputSchemaCompatibilityRows(ctx, inputSchemaID)
	if err != nil {
		return nil, err
	}
	result := make([]InputSchemaCompatibilityRowResponse, 0, len(rows))
	for _, row := range rows {
		name := row.WorkflowVersionName.String
		if !row.WorkflowVersionName.Valid || strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("Version %d", row.WorkflowVersionNumber)
		}
		result = append(result, InputSchemaCompatibilityRowResponse{
			ID:                    row.ID,
			WorkflowVersionID:     row.WorkflowVersionID,
			WorkflowVersionName:   name,
			WorkflowVersionNumber: row.WorkflowVersionNumber,
			WorkflowInputSchemaID: row.WorkflowInputSchemaID,
			CompatibilityType:     row.CompatibilityType,
			WorkflowInputMapperID: pgInt4ValuePtr(row.WorkflowInputMapperID),
			IsActive:              row.IsActive,
			IsDefaultRoute:        row.IsDefaultRoute,
		})
	}
	return result, nil
}

func (s *Service) CreateWorkflowInputSchemaRecord(ctx context.Context, workflowID int32, actorUserID int32, req inputSchemaMutation) (db.WorkflowInputSchema, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: code is required", ErrWorkflowConfigurationInvalid)
	}
	status := normalizeInputSchemaStatus(req.Status)
	if status == "" {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: unsupported input schema status", ErrWorkflowConfigurationInvalid)
	}
	isDefault := boolValue(req.IsDefault, false)
	if isDefault && status != "active" {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: default input schema must be active", ErrWorkflowConfigurationInvalid)
	}
	schemaJSON, err := normalizeJSONObject(req.SchemaJSON, []byte(`{"type":"object","properties":{}}`))
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}

	var created db.WorkflowInputSchema
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		versionNumber := req.VersionNumber
		if versionNumber <= 0 {
			next, err := tx.GetNextWorkflowInputSchemaVersionNumber(ctx, workflowID)
			if err != nil {
				return err
			}
			versionNumber = next
		}
		if isDefault {
			if err := tx.ClearDefaultWorkflowInputSchema(ctx, db.ClearDefaultWorkflowInputSchemaParams{
				WorkflowID:      workflowID,
				UpdatedByUserID: pgInt4(actorUserID),
			}); err != nil {
				return err
			}
		}
		var err error
		created, err = tx.CreateWorkflowInputSchema(ctx, db.CreateWorkflowInputSchemaParams{
			WorkflowID:      workflowID,
			Code:            code,
			VersionNumber:   versionNumber,
			SchemaJson:      schemaJSON,
			Status:          status,
			IsDefault:       isDefault,
			CreatedByUserID: pgInt4(actorUserID),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", created.ID, "create", actorUserID, nil, created)
	})
	if err != nil {
		return created, err
	}
	return created, nil
}

func (s *Service) UpdateWorkflowInputSchemaRecord(ctx context.Context, inputSchemaID int32, actorUserID int32, req inputSchemaMutation) (db.WorkflowInputSchema, error) {
	q, err := s.inputSchemaQueries()
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	existing, err := q.GetWorkflowInputSchemaByID(ctx, inputSchemaID)
	if err != nil {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	usage, err := s.GetWorkflowInputSchemaUsage(ctx, inputSchemaID)
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	if usage.IsReadOnly {
		return db.WorkflowInputSchema{}, ErrWorkflowConfigurationInUse
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = existing.Code
	}
	versionNumber := req.VersionNumber
	if versionNumber <= 0 {
		versionNumber = existing.VersionNumber
	}
	schemaJSON := existing.SchemaJson
	if len(bytes.TrimSpace(req.SchemaJSON)) > 0 {
		schemaJSON, err = normalizeJSONObject(req.SchemaJSON, nil)
		if err != nil {
			return db.WorkflowInputSchema{}, err
		}
	}

	var updated db.WorkflowInputSchema
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		updated, err = tx.UpdateWorkflowInputSchemaRecord(ctx, db.UpdateWorkflowInputSchemaRecordParams{
			ID:              inputSchemaID,
			Code:            code,
			VersionNumber:   versionNumber,
			SchemaJson:      schemaJSON,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		if boolValue(req.IsDefault, false) {
			if updated.Status != "active" {
				return fmt.Errorf("%w: default input schema must be active", ErrWorkflowConfigurationInvalid)
			}
			if err := tx.ClearDefaultWorkflowInputSchema(ctx, db.ClearDefaultWorkflowInputSchemaParams{
				WorkflowID:      updated.WorkflowID,
				UpdatedByUserID: pgInt4(actorUserID),
			}); err != nil {
				return err
			}
			updated, err = tx.SetDefaultWorkflowInputSchema(ctx, db.SetDefaultWorkflowInputSchemaParams{
				ID:              inputSchemaID,
				UpdatedByUserID: pgInt4(actorUserID),
			})
			if err != nil {
				return err
			}
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", updated.ID, "update", actorUserID, existing, updated)
	})
	if err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) UpdateWorkflowInputSchemaRecordStatus(ctx context.Context, inputSchemaID int32, actorUserID int32, status string) (db.WorkflowInputSchema, error) {
	if _, err := s.inputSchemaQueries(); err != nil {
		return db.WorkflowInputSchema{}, err
	}
	status = normalizeInputSchemaStatus(status)
	if status == "" {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: unsupported input schema status", ErrWorkflowConfigurationInvalid)
	}
	if status == "archived" {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: use archive endpoint for archived schemas", ErrWorkflowConfigurationInvalid)
	}
	usage, err := s.GetWorkflowInputSchemaUsage(ctx, inputSchemaID)
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	if usage.IsReadOnly {
		return db.WorkflowInputSchema{}, ErrWorkflowConfigurationInUse
	}
	var schema db.WorkflowInputSchema
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		schema, err = tx.UpdateWorkflowInputSchemaStatus(ctx, db.UpdateWorkflowInputSchemaStatusParams{
			ID:              inputSchemaID,
			Status:          status,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", schema.ID, "status", actorUserID, nil, schema)
	})
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	return schema, nil
}

func (s *Service) SetDefaultWorkflowInputSchemaRecord(ctx context.Context, inputSchemaID int32, actorUserID int32) (db.WorkflowInputSchema, error) {
	usage, err := s.GetWorkflowInputSchemaUsage(ctx, inputSchemaID)
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	if usage.IsReadOnly {
		return db.WorkflowInputSchema{}, ErrWorkflowConfigurationInUse
	}
	var updated db.WorkflowInputSchema
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		existing, err := tx.GetWorkflowInputSchemaByID(ctx, inputSchemaID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
		}
		if existing.Status != "active" {
			return fmt.Errorf("%w: default input schema must be active", ErrWorkflowConfigurationInvalid)
		}
		if err := tx.ClearDefaultWorkflowInputSchema(ctx, db.ClearDefaultWorkflowInputSchemaParams{
			WorkflowID:      existing.WorkflowID,
			UpdatedByUserID: pgInt4(actorUserID),
		}); err != nil {
			return err
		}
		updated, err = tx.SetDefaultWorkflowInputSchema(ctx, db.SetDefaultWorkflowInputSchemaParams{
			ID:              inputSchemaID,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", updated.ID, "set_default", actorUserID, nil, updated)
	})
	if err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) DeleteWorkflowInputSchemaRecord(ctx context.Context, inputSchemaID int32, actorUserID int32) error {
	q, err := s.inputSchemaQueries()
	if err != nil {
		return err
	}
	used, err := q.HasInputSchemaUsage(ctx, pgInt4(inputSchemaID))
	if err != nil {
		return err
	}
	if used {
		return s.store.WithTx(ctx, func(tx *db.Queries) error {
			schema, err := tx.UpdateWorkflowInputSchemaStatus(ctx, db.UpdateWorkflowInputSchemaStatusParams{
				ID:              inputSchemaID,
				Status:          "archived",
				UpdatedByUserID: pgInt4(actorUserID),
			})
			if err != nil {
				return err
			}
			return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", inputSchemaID, "archive", actorUserID, nil, schema)
		})
	}
	return s.store.WithTx(ctx, func(tx *db.Queries) error {
		if err := tx.SoftDeleteWorkflowInputSchema(ctx, inputSchemaID); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", inputSchemaID, "delete", actorUserID, nil, nil)
	})
}

func (s *Service) ArchiveWorkflowInputSchemaRecord(ctx context.Context, inputSchemaID int32, actorUserID int32, confirmationName string) (db.WorkflowInputSchema, error) {
	q, err := s.inputSchemaQueries()
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	existing, err := q.GetWorkflowInputSchemaByID(ctx, inputSchemaID)
	if err != nil {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	required := fmt.Sprintf("%s v%d", existing.Code, existing.VersionNumber)
	if strings.TrimSpace(confirmationName) != required {
		return db.WorkflowInputSchema{}, fmt.Errorf("%w: confirmation must match %q", ErrWorkflowConfigurationInvalid, required)
	}

	var archived db.WorkflowInputSchema
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		archived, err = tx.ArchiveWorkflowInputSchema(ctx, db.ArchiveWorkflowInputSchemaParams{
			ID:              inputSchemaID,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		if err := tx.DeactivateCompatibilitiesByInputSchemaID(ctx, db.DeactivateCompatibilitiesByInputSchemaIDParams{
			WorkflowInputSchemaID: inputSchemaID,
			UpdatedByUserID:       pgInt4(actorUserID),
		}); err != nil {
			return err
		}
		if err := tx.DeactivateMappedCompatibilitiesByNativeTargetSchemaID(ctx, db.DeactivateMappedCompatibilitiesByNativeTargetSchemaIDParams{
			WorkflowInputSchemaID: inputSchemaID,
			UpdatedByUserID:       pgInt4(actorUserID),
		}); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_schema", inputSchemaID, "archive", actorUserID, existing, archived)
	})
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	return archived, nil
}

func (s *Service) ListWorkflowInputMappers(ctx context.Context, workflowID int32) ([]db.WorkflowInputMapper, error) {
	q, err := s.inputMapperQueries()
	if err != nil {
		return nil, err
	}
	return q.ListWorkflowInputMappersByWorkflowID(ctx, workflowID)
}

func (s *Service) CreateWorkflowInputMapperRecord(ctx context.Context, workflowID int32, actorUserID int32, req inputMapperMutation) (db.WorkflowInputMapper, error) {
	if _, err := s.inputMapperQueries(); err != nil {
		return db.WorkflowInputMapper{}, err
	}
	mapperType := strings.TrimSpace(req.MapperType)
	if mapperType == "" {
		mapperType = "internal"
	}
	rules, err := normalizeJSON(req.Rules, []byte(`{}`))
	if err != nil {
		return db.WorkflowInputMapper{}, err
	}
	if err := validateWorkflowInputMapperRules(rules); err != nil {
		return db.WorkflowInputMapper{}, err
	}
	var mapper db.WorkflowInputMapper
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		mapper, err = tx.CreateWorkflowInputMapper(ctx, db.CreateWorkflowInputMapperParams{
			WorkflowID:      workflowID,
			MapperType:      mapperType,
			Rules:           rules,
			IsActive:        boolValue(req.IsActive, true),
			CreatedByUserID: pgInt4(actorUserID),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_mapper", mapper.ID, "create", actorUserID, nil, mapper)
	})
	if err != nil {
		return db.WorkflowInputMapper{}, err
	}
	return mapper, nil
}

func (s *Service) UpdateWorkflowInputMapperRecord(ctx context.Context, mapperID int32, actorUserID int32, req inputMapperMutation) (db.WorkflowInputMapper, error) {
	q, err := s.inputMapperQueries()
	if err != nil {
		return db.WorkflowInputMapper{}, err
	}
	existing, err := q.GetWorkflowInputMapperByID(ctx, mapperID)
	if err != nil {
		return db.WorkflowInputMapper{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	mapperType := strings.TrimSpace(req.MapperType)
	if mapperType == "" {
		mapperType = existing.MapperType
	}
	rules := existing.Rules
	if len(bytes.TrimSpace(req.Rules)) > 0 {
		rules, err = normalizeJSON(req.Rules, nil)
		if err != nil {
			return db.WorkflowInputMapper{}, err
		}
		if err := validateWorkflowInputMapperRules(rules); err != nil {
			return db.WorkflowInputMapper{}, err
		}
	}
	var mapper db.WorkflowInputMapper
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		mapper, err = tx.UpdateWorkflowInputMapper(ctx, db.UpdateWorkflowInputMapperParams{
			ID:              mapperID,
			MapperType:      mapperType,
			Rules:           rules,
			IsActive:        boolValue(req.IsActive, existing.IsActive),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_mapper", mapper.ID, "update", actorUserID, existing, mapper)
	})
	if err != nil {
		return db.WorkflowInputMapper{}, err
	}
	return mapper, nil
}

func (s *Service) UpdateWorkflowInputMapperActive(ctx context.Context, mapperID int32, actorUserID int32, isActive bool) (db.WorkflowInputMapper, error) {
	if _, err := s.inputMapperQueries(); err != nil {
		return db.WorkflowInputMapper{}, err
	}
	var mapper db.WorkflowInputMapper
	err := s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		mapper, err = tx.UpdateWorkflowInputMapperActive(ctx, db.UpdateWorkflowInputMapperActiveParams{
			ID:              mapperID,
			IsActive:        isActive,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_mapper", mapper.ID, "active", actorUserID, nil, mapper)
	})
	if err != nil {
		return db.WorkflowInputMapper{}, err
	}
	return mapper, nil
}

func (s *Service) DeleteWorkflowInputMapperRecord(ctx context.Context, mapperID int32, actorUserID int32) error {
	if _, err := s.inputMapperQueries(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(tx *db.Queries) error {
		if err := tx.SoftDeleteWorkflowInputMapper(ctx, mapperID); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_mapper", mapperID, "delete", actorUserID, nil, nil)
	})
}

func (s *Service) ListWorkflowCompatibilities(ctx context.Context, workflowID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error) {
	q, err := s.compatibilityQueries()
	if err != nil {
		return nil, err
	}
	return q.ListCompatibilitiesByWorkflowID(ctx, workflowID)
}

func (s *Service) ListVersionCompatibilities(ctx context.Context, versionID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error) {
	q, err := s.compatibilityQueries()
	if err != nil {
		return nil, err
	}
	return q.ListCompatibilitiesByVersionID(ctx, versionID)
}

func (s *Service) ListInputSchemaCompatibilities(ctx context.Context, inputSchemaID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error) {
	q, err := s.compatibilityQueries()
	if err != nil {
		return nil, err
	}
	return q.ListCompatibilitiesByInputSchemaID(ctx, inputSchemaID)
}

func (s *Service) CreateWorkflowCompatibility(ctx context.Context, versionID int32, actorUserID int32, req compatibilityMutation) (db.WorkflowVersionInputSchemaCompatibility, error) {
	if req.WorkflowInputSchemaID <= 0 {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: workflow_input_schema_id is required", ErrWorkflowConfigurationInvalid)
	}
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	schemaStore, err := s.inputSchemaQueries()
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	schema, err := schemaStore.GetWorkflowInputSchemaByID(ctx, req.WorkflowInputSchemaID)
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	if schema.WorkflowID != version.WorkflowID {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: input schema belongs to another workflow", ErrWorkflowConfigurationInvalid)
	}
	if err := s.ensureMapperBelongsToWorkflow(ctx, req.WorkflowInputMapperID, version.WorkflowID); err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}

	defaultValues, err := normalizeJSON(req.DefaultValues, []byte(`{}`))
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	compatibilityType := strings.TrimSpace(req.CompatibilityType)
	if compatibilityType == "" {
		compatibilityType = "native"
	}
	if err := validateCompatibilityMutationConfig(compatibilityType, req.WorkflowInputMapperID, defaultValues); err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	isDefaultRoute := boolValue(req.IsDefaultRoute, false)
	isActive := boolValue(req.IsActive, true)

	var created db.WorkflowVersionInputSchemaCompatibility
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		if isDefaultRoute {
			if err := clearCompatibilityDefaultRoute(ctx, tx, req.WorkflowInputSchemaID, actorUserID, 0); err != nil {
				return err
			}
		}
		var err error
		created, err = tx.CreateWorkflowVersionInputSchemaCompatibility(ctx, db.CreateWorkflowVersionInputSchemaCompatibilityParams{
			WorkflowVersionID:     versionID,
			WorkflowInputSchemaID: req.WorkflowInputSchemaID,
			CompatibilityType:     compatibilityType,
			WorkflowInputMapperID: pgInt4Ptr(req.WorkflowInputMapperID),
			DefaultValues:         defaultValues,
			IsActive:              isActive,
			IsDefaultRoute:        isDefaultRoute,
			CreatedByUserID:       pgInt4(actorUserID),
			UpdatedByUserID:       pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", created.ID, "create", actorUserID, nil, created)
	})
	if err != nil {
		return created, err
	}
	return created, nil
}

func (s *Service) ValidateAndCreateInputSchemaCompatibility(ctx context.Context, sourceSchemaID int32, actorUserID int32, req ValidateAndCreateCompatibilityRequest) (db.WorkflowVersionInputSchemaCompatibility, CompatibilityValidationResponse, error) {
	compatibility := db.WorkflowVersionInputSchemaCompatibility{}
	validation, mapperRules, targetVersionID, err := s.validateCompatibilityBuilderRequest(ctx, sourceSchemaID, req)
	if err != nil {
		return compatibility, validation, err
	}
	if !validation.IsValid {
		return compatibility, validation, nil
	}

	schemaQueries, err := s.inputSchemaQueries()
	if err != nil {
		return compatibility, validation, err
	}
	sourceSchema, err := schemaQueries.GetWorkflowInputSchemaByID(ctx, sourceSchemaID)
	if err != nil {
		return compatibility, validation, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		mapper, err := tx.CreateWorkflowInputMapper(ctx, db.CreateWorkflowInputMapperParams{
			WorkflowID:      sourceSchema.WorkflowID,
			MapperType:      "internal",
			Rules:           mapperRules,
			IsActive:        true,
			CreatedByUserID: pgInt4(actorUserID),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		if req.IsDefaultRoute {
			if err := clearCompatibilityDefaultRoute(ctx, tx, sourceSchemaID, actorUserID, 0); err != nil {
				return err
			}
		}
		compatibility, err = tx.CreateWorkflowVersionInputSchemaCompatibility(ctx, db.CreateWorkflowVersionInputSchemaCompatibilityParams{
			WorkflowVersionID:     targetVersionID,
			WorkflowInputSchemaID: sourceSchemaID,
			CompatibilityType:     "adapter",
			WorkflowInputMapperID: pgInt4(mapper.ID),
			DefaultValues:         []byte(`{}`),
			IsActive:              true,
			IsDefaultRoute:        req.IsDefaultRoute,
			CreatedByUserID:       pgInt4(actorUserID),
			UpdatedByUserID:       pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		if err := auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_input_mapper", mapper.ID, "create", actorUserID, nil, mapper); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", compatibility.ID, "create", actorUserID, nil, compatibility)
	})
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, validation, err
	}
	return compatibility, validation, nil
}

func (s *Service) UpdateWorkflowCompatibility(ctx context.Context, compatibilityID int32, actorUserID int32, req compatibilityMutation) (db.WorkflowVersionInputSchemaCompatibility, error) {
	q, err := s.compatibilityQueries()
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	existing, err := q.GetWorkflowVersionInputSchemaCompatibilityByID(ctx, compatibilityID)
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	version, err := s.store.GetWorkflowVersionByID(ctx, existing.WorkflowVersionID)
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	mapperID := req.WorkflowInputMapperID
	if mapperID == nil && existing.WorkflowInputMapperID.Valid {
		existingMapperID := existing.WorkflowInputMapperID.Int32
		mapperID = &existingMapperID
	}
	if err := s.ensureMapperBelongsToWorkflow(ctx, mapperID, version.WorkflowID); err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}

	compatibilityType := strings.TrimSpace(req.CompatibilityType)
	if compatibilityType == "" {
		compatibilityType = existing.CompatibilityType
	}
	defaultValues := existing.DefaultValues
	if len(bytes.TrimSpace(req.DefaultValues)) > 0 {
		defaultValues, err = normalizeJSON(req.DefaultValues, nil)
		if err != nil {
			return db.WorkflowVersionInputSchemaCompatibility{}, err
		}
	}
	if err := validateCompatibilityMutationConfig(compatibilityType, mapperID, defaultValues); err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	isDefaultRoute := boolValue(req.IsDefaultRoute, existing.IsDefaultRoute)
	isActive := boolValue(req.IsActive, existing.IsActive)

	var updated db.WorkflowVersionInputSchemaCompatibility
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		if isDefaultRoute {
			if err := clearCompatibilityDefaultRoute(ctx, tx, existing.WorkflowInputSchemaID, actorUserID, compatibilityID); err != nil {
				return err
			}
		}
		var err error
		updated, err = tx.UpdateWorkflowVersionInputSchemaCompatibility(ctx, db.UpdateWorkflowVersionInputSchemaCompatibilityParams{
			ID:                    compatibilityID,
			CompatibilityType:     compatibilityType,
			WorkflowInputMapperID: pgInt4Ptr(mapperID),
			DefaultValues:         defaultValues,
			IsActive:              isActive,
			IsDefaultRoute:        isDefaultRoute,
			UpdatedByUserID:       pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", updated.ID, "update", actorUserID, existing, updated)
	})
	if err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) UpdateWorkflowCompatibilityDefaultRoute(ctx context.Context, compatibilityID int32, actorUserID int32, isDefaultRoute bool) (db.WorkflowVersionInputSchemaCompatibility, error) {
	q, err := s.compatibilityQueries()
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	existing, err := q.GetWorkflowVersionInputSchemaCompatibilityByID(ctx, compatibilityID)
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}

	var updated db.WorkflowVersionInputSchemaCompatibility
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		if isDefaultRoute {
			if err := clearCompatibilityDefaultRoute(ctx, tx, existing.WorkflowInputSchemaID, actorUserID, compatibilityID); err != nil {
				return err
			}
		}
		var err error
		updated, err = tx.UpdateCompatibilityDefaultRoute(ctx, db.UpdateCompatibilityDefaultRouteParams{
			ID:              compatibilityID,
			IsDefaultRoute:  isDefaultRoute,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", updated.ID, "default_route", actorUserID, existing, updated)
	})
	if err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) DeactivateWorkflowCompatibility(ctx context.Context, compatibilityID int32, actorUserID int32) (db.WorkflowVersionInputSchemaCompatibility, error) {
	if _, err := s.compatibilityQueries(); err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	var compatibility db.WorkflowVersionInputSchemaCompatibility
	err := s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		compatibility, err = tx.DeactivateWorkflowVersionInputSchemaCompatibility(ctx, db.DeactivateWorkflowVersionInputSchemaCompatibilityParams{
			ID:              compatibilityID,
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", compatibility.ID, "deactivate", actorUserID, nil, compatibility)
	})
	if err != nil {
		return db.WorkflowVersionInputSchemaCompatibility{}, err
	}
	return compatibility, nil
}

func (s *Service) DeleteWorkflowCompatibility(ctx context.Context, compatibilityID int32, actorUserID int32) error {
	q, err := s.compatibilityQueries()
	if err != nil {
		return err
	}
	used, err := q.HasCompatibilityHistoricalRun(ctx, compatibilityID)
	if err != nil {
		return err
	}
	if used {
		return s.store.WithTx(ctx, func(tx *db.Queries) error {
			compatibility, err := tx.DeactivateWorkflowVersionInputSchemaCompatibility(ctx, db.DeactivateWorkflowVersionInputSchemaCompatibilityParams{
				ID:              compatibilityID,
				UpdatedByUserID: pgInt4(actorUserID),
			})
			if err != nil {
				return err
			}
			return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", compatibilityID, "deactivate", actorUserID, nil, compatibility)
		})
	}
	return s.store.WithTx(ctx, func(tx *db.Queries) error {
		if err := tx.SoftDeleteWorkflowVersionInputSchemaCompatibility(ctx, compatibilityID); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_version_input_schema_compatibility", compatibilityID, "delete", actorUserID, nil, nil)
	})
}

func (s *Service) ensureMapperBelongsToWorkflow(ctx context.Context, mapperID *int32, workflowID int32) error {
	if mapperID == nil || *mapperID <= 0 {
		return nil
	}
	q, err := s.inputMapperQueries()
	if err != nil {
		return err
	}
	mapper, err := q.GetWorkflowInputMapperByID(ctx, *mapperID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	if mapper.WorkflowID != workflowID {
		return fmt.Errorf("%w: input mapper belongs to another workflow", ErrWorkflowConfigurationInvalid)
	}
	return nil
}

func clearCompatibilityDefaultRoute(ctx context.Context, q workflowCompatibilityQueries, inputSchemaID int32, actorUserID int32, exceptID int32) error {
	existing, err := q.ListCompatibilitiesByInputSchemaID(ctx, inputSchemaID)
	if err != nil {
		return err
	}
	for _, compatibility := range existing {
		if compatibility.ID == exceptID || !compatibility.IsDefaultRoute {
			continue
		}
		if _, err := q.UpdateCompatibilityDefaultRoute(ctx, db.UpdateCompatibilityDefaultRouteParams{
			ID:              compatibility.ID,
			IsDefaultRoute:  false,
			UpdatedByUserID: pgInt4(actorUserID),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) validateCompatibilityBuilderRequest(ctx context.Context, sourceSchemaID int32, req ValidateAndCreateCompatibilityRequest) (CompatibilityValidationResponse, []byte, int32, error) {
	validation := CompatibilityValidationResponse{IsValid: true}
	schemaQueries, err := s.inputSchemaQueries()
	if err != nil {
		return validation, nil, 0, err
	}
	targetQueries, ok := s.store.(workflowCompatibilityTargetQueries)
	if !ok {
		return validation, nil, 0, ErrWorkflowConfigurationUnsupported
	}
	sourceSchema, err := schemaQueries.GetWorkflowInputSchemaByID(ctx, sourceSchemaID)
	if err != nil {
		return validation, nil, 0, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	targetSchema, err := schemaQueries.GetWorkflowInputSchemaByID(ctx, req.TargetInputSchemaID)
	if err != nil {
		return validation, nil, 0, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}

	var details []response.ErrorDetail
	if sourceSchema.WorkflowID != targetSchema.WorkflowID {
		details = append(details, compatibilityFieldError("", "source and target schemas must belong to the same workflow"))
	}
	if sourceSchema.Status == "archived" {
		details = append(details, compatibilityFieldError("", "source schema is archived"))
	}
	if targetSchema.Status == "archived" {
		details = append(details, compatibilityFieldError("", "target schema is archived"))
	}

	targetVersion, err := targetQueries.GetWorkflowVersionByNativeInputSchemaID(ctx, targetSchema.ID)
	if err != nil {
		details = append(details, compatibilityFieldError("target_input_schema_id", "target schema is not native for an available process version"))
	}
	if targetVersion.ArchivedAt.Valid {
		details = append(details, compatibilityFieldError("target_input_schema_id", "target process version is archived"))
	}

	sourceFields, err := parseWorkflowInputSchema(sourceSchema.SchemaJson)
	if err != nil {
		details = append(details, compatibilityFieldError("", fmt.Sprintf("source schema is invalid: %v", err)))
	}
	targetFields, err := parseWorkflowInputSchema(targetSchema.SchemaJson)
	if err != nil {
		details = append(details, compatibilityFieldError("", fmt.Sprintf("target schema is invalid: %v", err)))
	}

	mapping := map[string]any{}
	defaults := map[string]any{}
	ignored := make([]string, 0)
	seenTargets := map[string]struct{}{}
	for _, field := range req.Fields {
		targetPath := cleanCompatibilityFieldPath(field.TargetPath)
		sourcePath := cleanCompatibilityFieldPath(field.SourcePath)
		if targetPath == "" {
			details = append(details, compatibilityFieldError("target_path", "target path is required"))
			continue
		}
		if _, exists := seenTargets[targetPath]; exists {
			details = append(details, compatibilityFieldError(targetPath, "duplicate target path"))
			continue
		}
		seenTargets[targetPath] = struct{}{}
		targetField, ok := targetFields[targetPath]
		if !ok {
			details = append(details, compatibilityFieldError(targetPath, "target field is not declared"))
			continue
		}

		hasSource := sourcePath != ""
		hasDefault := len(bytes.TrimSpace(field.Default)) > 0
		hasIgnore := field.Ignore
		actionCount := 0
		for _, active := range []bool{hasSource, hasDefault, hasIgnore} {
			if active {
				actionCount++
			}
		}
		if actionCount != 1 {
			details = append(details, compatibilityFieldError(targetPath, "target field must have exactly one source, default, or ignore action"))
			continue
		}
		switch {
		case hasSource:
			sourceField, ok := sourceFields[sourcePath]
			if !ok {
				details = append(details, compatibilityFieldError(targetPath, "source field is not declared"))
				continue
			}
			if sourceField.Type != "" && targetField.Type != "" && sourceField.Type != targetField.Type {
				details = append(details, compatibilityFieldError(targetPath, fmt.Sprintf("type mismatch: source %s, target %s", sourceField.Type, targetField.Type)))
				continue
			}
			mapping[targetPath] = sourcePath
		case hasDefault:
			var value any
			if err := json.Unmarshal(field.Default, &value); err != nil {
				details = append(details, compatibilityFieldError(targetPath, "default value must be valid JSON"))
				continue
			}
			if err := validateDefaultValueForSchemaType(value, targetField.Type); err != nil {
				details = append(details, compatibilityFieldError(targetPath, err.Error()))
				continue
			}
			defaults[targetPath] = value
		case hasIgnore:
			ignored = append(ignored, targetPath)
		}
	}
	for targetPath := range targetFields {
		if _, ok := seenTargets[targetPath]; !ok {
			details = append(details, compatibilityFieldError(targetPath, "target field is not covered"))
		}
	}

	if len(details) > 0 {
		validation.IsValid = false
		validation.Errors = details
		return validation, nil, 0, nil
	}
	rules, err := json.Marshal(map[string]any{
		"copy_all": false,
		"mapping":  mapping,
		"defaults": defaults,
		"ignore":   ignored,
	})
	if err != nil {
		return validation, nil, 0, err
	}
	return validation, rules, targetVersion.ID, nil
}

func decodeRoutingSupportedSchemas(value any) ([]RoutingSupportedSchemaResponse, error) {
	if value == nil {
		return []RoutingSupportedSchemaResponse{}, nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		raw = encoded
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return []RoutingSupportedSchemaResponse{}, nil
	}
	var supported []RoutingSupportedSchemaResponse
	if err := json.Unmarshal(raw, &supported); err != nil {
		return nil, err
	}
	if supported == nil {
		return []RoutingSupportedSchemaResponse{}, nil
	}
	return supported, nil
}

func compatibilityFieldError(field string, message string) response.ErrorDetail {
	return response.ErrorDetail{
		Type:    "compatibility_validation",
		Field:   field,
		Message: message,
	}
}

func cleanCompatibilityFieldPath(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "$.")
	value = strings.TrimPrefix(value, "payload.")
	return value
}

func validateDefaultValueForSchemaType(value any, schemaType string) error {
	switch schemaType {
	case "", "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("default value must be a string")
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("default value must be a number")
		}
	case "integer":
		number, ok := value.(float64)
		if !ok || number != float64(int64(number)) {
			return fmt.Errorf("default value must be an integer")
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("default value must be a boolean")
		}
	}
	return nil
}

func pgTextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func pgInt4ValuePtr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func validateCompatibilityMutationConfig(compatibilityType string, mapperID *int32, defaultValues []byte) error {
	hasMapper := mapperID != nil && *mapperID > 0
	hasDefaults := hasCompatibilityDefaultValues(defaultValues)
	switch compatibilityType {
	case "native":
		if hasMapper || hasDefaults {
			return fmt.Errorf("%w: native compatibility must not define mapper or default values", ErrWorkflowConfigurationInvalid)
		}
	case "adapter", "partial":
		if !hasMapper && !hasDefaults {
			return fmt.Errorf("%w: %s compatibility requires mapper or default values", ErrWorkflowConfigurationInvalid, compatibilityType)
		}
	default:
		return fmt.Errorf("%w: unsupported compatibility type %q", ErrWorkflowConfigurationInvalid, compatibilityType)
	}
	return nil
}

func normalizeInputSchemaStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "draft"
	}
	switch status {
	case "draft", "active", "deprecated", "archived":
		return status
	default:
		return ""
	}
}

func normalizeJSON(raw json.RawMessage, defaultValue []byte) ([]byte, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return append([]byte(nil), defaultValue...), nil
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("%w: invalid json", ErrWorkflowConfigurationInvalid)
	}
	return append([]byte(nil), raw...), nil
}

func normalizeJSONObject(raw json.RawMessage, defaultValue []byte) ([]byte, error) {
	value, err := normalizeJSON(raw, defaultValue)
	if err != nil {
		return nil, err
	}
	value = bytes.TrimSpace(value)
	if len(value) == 0 || value[0] != '{' {
		return nil, fmt.Errorf("%w: json object is required", ErrWorkflowConfigurationInvalid)
	}
	return value, nil
}

func validateWorkflowInputMapperRules(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte(`{}`)) {
		return nil
	}

	var rules map[string]any
	if err := json.Unmarshal(raw, &rules); err != nil {
		return fmt.Errorf("%w: mapper rules must be a json object", ErrWorkflowConfigurationInvalid)
	}
	if rules == nil {
		return fmt.Errorf("%w: mapper rules must be a json object", ErrWorkflowConfigurationInvalid)
	}
	if copyAll, exists := rules["copy_all"]; exists {
		if _, ok := copyAll.(bool); !ok {
			return fmt.Errorf("%w: mapper copy_all must be boolean", ErrWorkflowConfigurationInvalid)
		}
	}
	if defaults, exists := rules["defaults"]; exists {
		if _, ok := defaults.(map[string]any); !ok {
			return fmt.Errorf("%w: mapper defaults must be an object", ErrWorkflowConfigurationInvalid)
		}
	}
	if ignore, exists := rules["ignore"]; exists {
		items, ok := ignore.([]any)
		if !ok {
			return fmt.Errorf("%w: mapper ignore must be an array", ErrWorkflowConfigurationInvalid)
		}
		for _, item := range items {
			path, ok := item.(string)
			if !ok || strings.TrimSpace(path) == "" {
				return fmt.Errorf("%w: mapper ignore entries must be non-empty strings", ErrWorkflowConfigurationInvalid)
			}
		}
	}

	for _, key := range []string{"fields", "mapping", "mappings"} {
		value, exists := rules[key]
		if !exists {
			continue
		}
		mappings, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: mapper %s must be an object", ErrWorkflowConfigurationInvalid, key)
		}
		for targetPath, sourceSpec := range mappings {
			if strings.TrimSpace(targetPath) == "" {
				return fmt.Errorf("%w: mapper target path is required", ErrWorkflowConfigurationInvalid)
			}
			if err := validateWorkflowInputMapperSourceSpec(sourceSpec); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateWorkflowInputMapperSourceSpec(sourceSpec any) error {
	switch value := sourceSpec.(type) {
	case string:
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: mapper source path is required", ErrWorkflowConfigurationInvalid)
		}
		return nil
	case map[string]any:
		if _, ok := value["literal"]; ok {
			return nil
		}
		if source, ok := value["source"].(string); ok && strings.TrimSpace(source) != "" {
			return nil
		}
		if path, ok := value["path"].(string); ok && strings.TrimSpace(path) != "" {
			return nil
		}
		return fmt.Errorf("%w: mapper source object must contain source, path, or literal", ErrWorkflowConfigurationInvalid)
	default:
		return fmt.Errorf("%w: mapper source must be a path string or object", ErrWorkflowConfigurationInvalid)
	}
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func pgInt4(value int32) pgtype.Int4 {
	if value <= 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: value, Valid: true}
}

func pgInt4Ptr(value *int32) pgtype.Int4 {
	if value == nil || *value <= 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}
