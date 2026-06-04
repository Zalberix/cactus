package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

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

type inputSchemaMutation struct {
	Code          string
	VersionNumber int32
	SchemaJSON    json.RawMessage
	Status        string
	IsDefault     *bool
}

type inputMapperMutation struct {
	Name       string
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
	used, err := q.HasInputSchemaUsage(ctx, pgInt4(inputSchemaID))
	if err != nil {
		return db.WorkflowInputSchema{}, err
	}
	if used {
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
	var schema db.WorkflowInputSchema
	err := s.store.WithTx(ctx, func(tx *db.Queries) error {
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
	var updated db.WorkflowInputSchema
	err := s.store.WithTx(ctx, func(tx *db.Queries) error {
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
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return db.WorkflowInputMapper{}, fmt.Errorf("%w: mapper name is required", ErrWorkflowConfigurationInvalid)
	}
	mapperType := strings.TrimSpace(req.MapperType)
	if mapperType == "" {
		mapperType = "json"
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
			Name:            name,
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
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
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
			Name:            name,
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
