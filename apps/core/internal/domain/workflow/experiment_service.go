package workflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type workflowExperimentQueries interface {
	CreateWorkflowExperiment(ctx context.Context, arg db.CreateWorkflowExperimentParams) (db.WorkflowExperiment, error)
	CreateWorkflowExperimentScope(ctx context.Context, arg db.CreateWorkflowExperimentScopeParams) (db.WorkflowExperimentScope, error)
	CreateWorkflowExperimentVariant(ctx context.Context, arg db.CreateWorkflowExperimentVariantParams) (db.WorkflowExperimentVariant, error)
	GetWorkflowExperimentByID(ctx context.Context, id int32) (db.WorkflowExperiment, error)
	GetWorkflowExperimentScopeByID(ctx context.Context, id int32) (db.WorkflowExperimentScope, error)
	GetWorkflowExperimentVariantByID(ctx context.Context, id int32) (db.WorkflowExperimentVariant, error)
	ListWorkflowExperimentScopesByExperimentID(ctx context.Context, workflowExperimentID int32) ([]db.WorkflowExperimentScope, error)
	ListWorkflowExperimentVariantsByScopeID(ctx context.Context, workflowExperimentScopeID int32) ([]db.WorkflowExperimentVariant, error)
	ListWorkflowExperimentsByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowExperiment, error)
	SoftDeleteWorkflowExperiment(ctx context.Context, id int32) error
	SoftDeleteWorkflowExperimentScope(ctx context.Context, id int32) error
	SoftDeleteWorkflowExperimentVariant(ctx context.Context, id int32) error
	UpdateWorkflowExperiment(ctx context.Context, arg db.UpdateWorkflowExperimentParams) (db.WorkflowExperiment, error)
	UpdateWorkflowExperimentScope(ctx context.Context, arg db.UpdateWorkflowExperimentScopeParams) (db.WorkflowExperimentScope, error)
	UpdateWorkflowExperimentStatus(ctx context.Context, arg db.UpdateWorkflowExperimentStatusParams) (db.WorkflowExperiment, error)
	UpdateWorkflowExperimentVariant(ctx context.Context, arg db.UpdateWorkflowExperimentVariantParams) (db.WorkflowExperimentVariant, error)
}

type workflowExperimentMutation struct {
	Name           string
	Description    string
	ExperimentType string
	Status         string
	StartedAt      *time.Time
	EndedAt        *time.Time
}

type workflowExperimentStatusMutation struct {
	Status    string
	StartedAt *time.Time
	EndedAt   *time.Time
}

type workflowExperimentScopeMutation struct {
	WorkflowInputSchemaID     int32
	TrafficConditions         json.RawMessage
	TrafficPercent            *int32
	FallbackPolicy            string
	FallbackWorkflowVersionID *int32
}

type workflowExperimentVariantMutation struct {
	WorkflowVersionID int32
	TrafficWeight     *int32
	IsControlGroup    *bool
	IsActive          *bool
}

func (s *Service) experimentQueries() (workflowExperimentQueries, error) {
	q, ok := s.store.(workflowExperimentQueries)
	if !ok {
		return nil, ErrWorkflowConfigurationUnsupported
	}
	return q, nil
}

func (s *Service) ListWorkflowExperiments(ctx context.Context, workflowID int32) ([]db.WorkflowExperiment, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return nil, err
	}
	return q.ListWorkflowExperimentsByWorkflowID(ctx, workflowID)
}

func (s *Service) GetWorkflowExperimentRecord(ctx context.Context, experimentID int32) (db.WorkflowExperiment, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return db.WorkflowExperiment{}, err
	}
	experiment, err := q.GetWorkflowExperimentByID(ctx, experimentID)
	if err != nil {
		return db.WorkflowExperiment{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	return experiment, nil
}

func (s *Service) CreateWorkflowExperimentRecord(ctx context.Context, workflowID int32, actorUserID int32, req workflowExperimentMutation) (db.WorkflowExperiment, error) {
	if _, err := s.experimentQueries(); err != nil {
		return db.WorkflowExperiment{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return db.WorkflowExperiment{}, fmt.Errorf("%w: experiment name is required", ErrWorkflowConfigurationInvalid)
	}
	experimentType := normalizeExperimentType(req.ExperimentType)
	if experimentType == "" {
		return db.WorkflowExperiment{}, fmt.Errorf("%w: unsupported experiment type", ErrWorkflowConfigurationInvalid)
	}
	status := normalizeExperimentStatus(req.Status)
	if status == "" {
		return db.WorkflowExperiment{}, fmt.Errorf("%w: unsupported experiment status", ErrWorkflowConfigurationInvalid)
	}
	if err := validateTimeRange(req.StartedAt, req.EndedAt); err != nil {
		return db.WorkflowExperiment{}, err
	}
	var experiment db.WorkflowExperiment
	err := s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		experiment, err = tx.CreateWorkflowExperiment(ctx, db.CreateWorkflowExperimentParams{
			WorkflowID:      workflowID,
			Name:            name,
			Description:     pgText(req.Description),
			ExperimentType:  experimentType,
			Status:          status,
			StartedAt:       pgTimestamp(req.StartedAt),
			EndedAt:         pgTimestamp(req.EndedAt),
			CreatedByUserID: pgInt4(actorUserID),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment", experiment.ID, "create", actorUserID, nil, experiment)
	})
	if err != nil {
		return db.WorkflowExperiment{}, err
	}
	return experiment, nil
}

func (s *Service) UpdateWorkflowExperimentRecord(ctx context.Context, experimentID int32, actorUserID int32, req workflowExperimentMutation) (db.WorkflowExperiment, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return db.WorkflowExperiment{}, err
	}
	existing, err := q.GetWorkflowExperimentByID(ctx, experimentID)
	if err != nil {
		return db.WorkflowExperiment{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}
	experimentType := normalizeExperimentType(req.ExperimentType)
	if experimentType == "" {
		experimentType = existing.ExperimentType
	}
	if err := validateTimeRange(req.StartedAt, req.EndedAt); err != nil {
		return db.WorkflowExperiment{}, err
	}
	var experiment db.WorkflowExperiment
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		experiment, err = tx.UpdateWorkflowExperiment(ctx, db.UpdateWorkflowExperimentParams{
			ID:              experimentID,
			Name:            name,
			Description:     pgText(req.Description),
			ExperimentType:  experimentType,
			StartedAt:       pgTimestamp(req.StartedAt),
			EndedAt:         pgTimestamp(req.EndedAt),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment", experiment.ID, "update", actorUserID, existing, experiment)
	})
	if err != nil {
		return db.WorkflowExperiment{}, err
	}
	return experiment, nil
}

func (s *Service) UpdateWorkflowExperimentRecordStatus(ctx context.Context, experimentID int32, actorUserID int32, req workflowExperimentStatusMutation) (db.WorkflowExperiment, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return db.WorkflowExperiment{}, err
	}
	status := normalizeExperimentStatus(req.Status)
	if status == "" {
		return db.WorkflowExperiment{}, fmt.Errorf("%w: unsupported experiment status", ErrWorkflowConfigurationInvalid)
	}
	if err := validateTimeRange(req.StartedAt, req.EndedAt); err != nil {
		return db.WorkflowExperiment{}, err
	}
	if status == "active" {
		if err := s.validateExperimentActivation(ctx, q, experimentID); err != nil {
			return db.WorkflowExperiment{}, err
		}
	}
	startedAt := req.StartedAt
	endedAt := req.EndedAt
	now := time.Now().UTC()
	if status == "active" && startedAt == nil {
		startedAt = &now
	}
	if (status == "completed" || status == "cancelled") && endedAt == nil {
		endedAt = &now
	}
	var experiment db.WorkflowExperiment
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		experiment, err = tx.UpdateWorkflowExperimentStatus(ctx, db.UpdateWorkflowExperimentStatusParams{
			ID:              experimentID,
			Status:          status,
			StartedAt:       pgTimestamp(startedAt),
			EndedAt:         pgTimestamp(endedAt),
			UpdatedByUserID: pgInt4(actorUserID),
		})
		if err != nil {
			return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment", experiment.ID, "status", actorUserID, nil, experiment)
	})
	if err != nil {
		return db.WorkflowExperiment{}, err
	}
	return experiment, nil
}

func (s *Service) validateExperimentActivation(ctx context.Context, q workflowExperimentQueries, experimentID int32) error {
	experiment, err := q.GetWorkflowExperimentByID(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	scopes, err := q.ListWorkflowExperimentScopesByExperimentID(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("list experiment scopes: %w", err)
	}
	if len(scopes) == 0 {
		return fmt.Errorf("%w: active experiment requires at least one scope", ErrWorkflowConfigurationInvalid)
	}
	schemaQueries, err := s.inputSchemaQueries()
	if err != nil {
		return err
	}
	compatQueries, err := s.compatibilityQueries()
	if err != nil {
		return err
	}
	for _, scope := range scopes {
		schema, err := schemaQueries.GetWorkflowInputSchemaByID(ctx, scope.WorkflowInputSchemaID)
		if err != nil {
			return fmt.Errorf("%w: input schema %d is not available", ErrWorkflowConfigurationInvalid, scope.WorkflowInputSchemaID)
		}
		if schema.WorkflowID != experiment.WorkflowID {
			return fmt.Errorf("%w: scope %d input schema belongs to another workflow", ErrWorkflowConfigurationInvalid, scope.ID)
		}
		if schema.Status != "active" {
			return fmt.Errorf("%w: scope %d input schema must be active", ErrWorkflowConfigurationInvalid, scope.ID)
		}

		variants, err := q.ListWorkflowExperimentVariantsByScopeID(ctx, scope.ID)
		if err != nil {
			return fmt.Errorf("list experiment variants for scope %d: %w", scope.ID, err)
		}
		var activeCount int
		var controlCount int
		var weightTotal int32
		for _, variant := range variants {
			if !variant.IsActive {
				continue
			}
			activeCount++
			if variant.IsControlGroup {
				controlCount++
			}
			weightTotal += variant.TrafficWeight
			if err := s.validateExperimentVersionForScope(ctx, compatQueries, experiment.WorkflowID, scope.WorkflowInputSchemaID, variant.WorkflowVersionID, "variant"); err != nil {
				return err
			}
		}
		if activeCount == 0 {
			return fmt.Errorf("%w: scope %d requires at least one active variant", ErrWorkflowConfigurationInvalid, scope.ID)
		}
		if weightTotal != 100 {
			return fmt.Errorf("%w: scope %d active variant traffic_weight total must equal 100", ErrWorkflowConfigurationInvalid, scope.ID)
		}
		if controlCount > 1 {
			return fmt.Errorf("%w: scope %d must not have more than one control variant", ErrWorkflowConfigurationInvalid, scope.ID)
		}
		if scope.FallbackWorkflowVersionID.Valid {
			if err := s.validateExperimentVersionForScope(ctx, compatQueries, experiment.WorkflowID, scope.WorkflowInputSchemaID, scope.FallbackWorkflowVersionID.Int32, "fallback"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) validateExperimentVersionForScope(ctx context.Context, compatQueries workflowCompatibilityQueries, workflowID int32, inputSchemaID int32, versionID int32, role string) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("%w: %s workflow version %d is not available", ErrWorkflowConfigurationInvalid, role, versionID)
	}
	if version.WorkflowID != workflowID {
		return fmt.Errorf("%w: %s workflow version %d belongs to another workflow", ErrWorkflowConfigurationInvalid, role, versionID)
	}
	if !version.IsActive || !version.IsValid || version.ArchivedAt.Valid {
		return fmt.Errorf("%w: %s workflow version %d must be active, valid, and not archived", ErrWorkflowConfigurationInvalid, role, versionID)
	}
	compatibilities, err := compatQueries.ListCompatibilitiesByVersionID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("list compatibilities for %s workflow version %d: %w", role, versionID, err)
	}
	for _, compatibility := range compatibilities {
		if compatibility.IsActive && compatibility.WorkflowInputSchemaID == inputSchemaID {
			return nil
		}
	}
	return fmt.Errorf("%w: %s workflow version %d has no active compatibility with input schema %d", ErrWorkflowConfigurationInvalid, role, versionID, inputSchemaID)
}

func (s *Service) DeleteWorkflowExperimentRecord(ctx context.Context, experimentID int32) error {
	if _, err := s.experimentQueries(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(tx *db.Queries) error {
		if err := tx.SoftDeleteWorkflowExperiment(ctx, experimentID); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment", experimentID, "delete", 0, nil, nil)
	})
}

func (s *Service) ListWorkflowExperimentScopes(ctx context.Context, experimentID int32) ([]db.WorkflowExperimentScope, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return nil, err
	}
	return q.ListWorkflowExperimentScopesByExperimentID(ctx, experimentID)
}

func (s *Service) CreateWorkflowExperimentScopeRecord(ctx context.Context, experimentID int32, actorUserID int32, req workflowExperimentScopeMutation) (db.WorkflowExperimentScope, error) {
	if req.WorkflowInputSchemaID <= 0 {
		return db.WorkflowExperimentScope{}, fmt.Errorf("%w: workflow_input_schema_id is required", ErrWorkflowConfigurationInvalid)
	}
	q, err := s.experimentQueries()
	if err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	experiment, err := q.GetWorkflowExperimentByID(ctx, experimentID)
	if err != nil {
		return db.WorkflowExperimentScope{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	if err := s.ensureInputSchemaBelongsToWorkflow(ctx, req.WorkflowInputSchemaID, experiment.WorkflowID); err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	if err := s.ensureVersionBelongsToWorkflow(ctx, req.FallbackWorkflowVersionID, experiment.WorkflowID); err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	conditions, hash, err := normalizeConditions(req.TrafficConditions)
	if err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	trafficPercent := int32Value(req.TrafficPercent, 100)
	if trafficPercent < 0 || trafficPercent > 100 {
		return db.WorkflowExperimentScope{}, fmt.Errorf("%w: traffic_percent must be between 0 and 100", ErrWorkflowConfigurationInvalid)
	}
	var scope db.WorkflowExperimentScope
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		scope, err = tx.CreateWorkflowExperimentScope(ctx, db.CreateWorkflowExperimentScopeParams{
			WorkflowExperimentID:      experimentID,
			WorkflowInputSchemaID:     req.WorkflowInputSchemaID,
			TrafficConditions:         conditions,
			ConditionsHash:            hash,
			TrafficPercent:            trafficPercent,
			FallbackPolicy:            normalizeFallbackPolicy(req.FallbackPolicy),
			FallbackWorkflowVersionID: pgInt4Ptr(req.FallbackWorkflowVersionID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment_scope", scope.ID, "create", actorUserID, nil, scope)
	})
	if err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	return scope, nil
}

func (s *Service) UpdateWorkflowExperimentScopeRecord(ctx context.Context, scopeID int32, actorUserID int32, req workflowExperimentScopeMutation) (db.WorkflowExperimentScope, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	existing, err := q.GetWorkflowExperimentScopeByID(ctx, scopeID)
	if err != nil {
		return db.WorkflowExperimentScope{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	experiment, err := q.GetWorkflowExperimentByID(ctx, existing.WorkflowExperimentID)
	if err != nil {
		return db.WorkflowExperimentScope{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	inputSchemaID := req.WorkflowInputSchemaID
	if inputSchemaID <= 0 {
		inputSchemaID = existing.WorkflowInputSchemaID
	}
	if err := s.ensureInputSchemaBelongsToWorkflow(ctx, inputSchemaID, experiment.WorkflowID); err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	fallbackVersionID := req.FallbackWorkflowVersionID
	if fallbackVersionID == nil && existing.FallbackWorkflowVersionID.Valid {
		value := existing.FallbackWorkflowVersionID.Int32
		fallbackVersionID = &value
	}
	if err := s.ensureVersionBelongsToWorkflow(ctx, fallbackVersionID, experiment.WorkflowID); err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	conditions := existing.TrafficConditions
	hash := existing.ConditionsHash
	if len(bytes.TrimSpace(req.TrafficConditions)) > 0 {
		conditions, hash, err = normalizeConditions(req.TrafficConditions)
		if err != nil {
			return db.WorkflowExperimentScope{}, err
		}
	}
	trafficPercent := int32Value(req.TrafficPercent, existing.TrafficPercent)
	if trafficPercent < 0 || trafficPercent > 100 {
		return db.WorkflowExperimentScope{}, fmt.Errorf("%w: traffic_percent must be between 0 and 100", ErrWorkflowConfigurationInvalid)
	}
	var scope db.WorkflowExperimentScope
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		var err error
		scope, err = tx.UpdateWorkflowExperimentScope(ctx, db.UpdateWorkflowExperimentScopeParams{
			ID:                        scopeID,
			WorkflowInputSchemaID:     inputSchemaID,
			TrafficConditions:         conditions,
			ConditionsHash:            hash,
			TrafficPercent:            trafficPercent,
			FallbackPolicy:            normalizeFallbackPolicy(req.FallbackPolicy),
			FallbackWorkflowVersionID: pgInt4Ptr(fallbackVersionID),
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment_scope", scope.ID, "update", actorUserID, existing, scope)
	})
	if err != nil {
		return db.WorkflowExperimentScope{}, err
	}
	return scope, nil
}

func (s *Service) DeleteWorkflowExperimentScopeRecord(ctx context.Context, scopeID int32, actorUserID int32) error {
	if _, err := s.experimentQueries(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(tx *db.Queries) error {
		if err := tx.SoftDeleteWorkflowExperimentScope(ctx, scopeID); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment_scope", scopeID, "delete", actorUserID, nil, nil)
	})
}

func (s *Service) ListWorkflowExperimentVariants(ctx context.Context, scopeID int32) ([]db.WorkflowExperimentVariant, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return nil, err
	}
	return q.ListWorkflowExperimentVariantsByScopeID(ctx, scopeID)
}

func (s *Service) CreateWorkflowExperimentVariantRecord(ctx context.Context, scopeID int32, actorUserID int32, req workflowExperimentVariantMutation) (db.WorkflowExperimentVariant, error) {
	if req.WorkflowVersionID <= 0 {
		return db.WorkflowExperimentVariant{}, fmt.Errorf("%w: workflow_version_id is required", ErrWorkflowConfigurationInvalid)
	}
	workflowID, err := s.workflowIDForExperimentScope(ctx, scopeID)
	if err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	if err := s.ensureVersionBelongsToWorkflow(ctx, &req.WorkflowVersionID, workflowID); err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	weight := int32Value(req.TrafficWeight, 100)
	if err := validateTrafficWeight(weight); err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	isActive := boolValue(req.IsActive, true)
	isControl := boolValue(req.IsControlGroup, false)
	var created db.WorkflowExperimentVariant
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		variants, err := tx.ListWorkflowExperimentVariantsByScopeID(ctx, scopeID)
		if err != nil {
			return err
		}
		if err := validateVariantWeightTotal(variants, 0, weight, isActive); err != nil {
			return err
		}
		if isControl {
			if err := clearControlGroupVariant(ctx, tx, variants, 0); err != nil {
				return err
			}
		}
		created, err = tx.CreateWorkflowExperimentVariant(ctx, db.CreateWorkflowExperimentVariantParams{
			WorkflowExperimentScopeID: scopeID,
			WorkflowVersionID:         req.WorkflowVersionID,
			TrafficWeight:             weight,
			IsControlGroup:            isControl,
			IsActive:                  isActive,
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment_variant", created.ID, "create", actorUserID, nil, created)
	})
	if err != nil {
		return created, err
	}
	return created, nil
}

func (s *Service) UpdateWorkflowExperimentVariantRecord(ctx context.Context, variantID int32, actorUserID int32, req workflowExperimentVariantMutation) (db.WorkflowExperimentVariant, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	existing, err := q.GetWorkflowExperimentVariantByID(ctx, variantID)
	if err != nil {
		return db.WorkflowExperimentVariant{}, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	workflowID, err := s.workflowIDForExperimentScope(ctx, existing.WorkflowExperimentScopeID)
	if err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	versionID := req.WorkflowVersionID
	if versionID <= 0 {
		versionID = existing.WorkflowVersionID
	}
	if err := s.ensureVersionBelongsToWorkflow(ctx, &versionID, workflowID); err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	weight := int32Value(req.TrafficWeight, existing.TrafficWeight)
	if err := validateTrafficWeight(weight); err != nil {
		return db.WorkflowExperimentVariant{}, err
	}
	isActive := boolValue(req.IsActive, existing.IsActive)
	isControl := boolValue(req.IsControlGroup, existing.IsControlGroup)
	var updated db.WorkflowExperimentVariant
	err = s.store.WithTx(ctx, func(tx *db.Queries) error {
		variants, err := tx.ListWorkflowExperimentVariantsByScopeID(ctx, existing.WorkflowExperimentScopeID)
		if err != nil {
			return err
		}
		if err := validateVariantWeightTotal(variants, variantID, weight, isActive); err != nil {
			return err
		}
		if isControl {
			if err := clearControlGroupVariant(ctx, tx, variants, variantID); err != nil {
				return err
			}
		}
		updated, err = tx.UpdateWorkflowExperimentVariant(ctx, db.UpdateWorkflowExperimentVariantParams{
			ID:                variantID,
			WorkflowVersionID: versionID,
			TrafficWeight:     weight,
			IsControlGroup:    isControl,
			IsActive:          isActive,
		})
		if err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment_variant", updated.ID, "update", actorUserID, existing, updated)
	})
	if err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) DeleteWorkflowExperimentVariantRecord(ctx context.Context, variantID int32, actorUserID int32) error {
	if _, err := s.experimentQueries(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(tx *db.Queries) error {
		if err := tx.SoftDeleteWorkflowExperimentVariant(ctx, variantID); err != nil {
			return err
		}
		return auditWorkflowConfigurationWithQueries(ctx, tx, "workflow_experiment_variant", variantID, "delete", actorUserID, nil, nil)
	})
}

func (s *Service) workflowIDForExperimentScope(ctx context.Context, scopeID int32) (int32, error) {
	q, err := s.experimentQueries()
	if err != nil {
		return 0, err
	}
	scope, err := q.GetWorkflowExperimentScopeByID(ctx, scopeID)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	experiment, err := q.GetWorkflowExperimentByID(ctx, scope.WorkflowExperimentID)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	return experiment.WorkflowID, nil
}

func (s *Service) ensureInputSchemaBelongsToWorkflow(ctx context.Context, inputSchemaID int32, workflowID int32) error {
	q, err := s.inputSchemaQueries()
	if err != nil {
		return err
	}
	schema, err := q.GetWorkflowInputSchemaByID(ctx, inputSchemaID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	if schema.WorkflowID != workflowID {
		return fmt.Errorf("%w: input schema belongs to another workflow", ErrWorkflowConfigurationInvalid)
	}
	return nil
}

func (s *Service) ensureVersionBelongsToWorkflow(ctx context.Context, versionID *int32, workflowID int32) error {
	if versionID == nil || *versionID <= 0 {
		return nil
	}
	version, err := s.store.GetWorkflowVersionByID(ctx, *versionID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWorkflowConfigurationNotFound, err)
	}
	if version.WorkflowID != workflowID {
		return fmt.Errorf("%w: workflow version belongs to another workflow", ErrWorkflowConfigurationInvalid)
	}
	return nil
}

func clearControlGroupVariant(ctx context.Context, q workflowExperimentQueries, variants []db.WorkflowExperimentVariant, exceptID int32) error {
	for _, variant := range variants {
		if variant.ID == exceptID || !variant.IsControlGroup {
			continue
		}
		if _, err := q.UpdateWorkflowExperimentVariant(ctx, db.UpdateWorkflowExperimentVariantParams{
			ID:                variant.ID,
			WorkflowVersionID: variant.WorkflowVersionID,
			TrafficWeight:     variant.TrafficWeight,
			IsControlGroup:    false,
			IsActive:          variant.IsActive,
		}); err != nil {
			return err
		}
	}
	return nil
}

func validateVariantWeightTotal(variants []db.WorkflowExperimentVariant, replacementID int32, replacementWeight int32, replacementActive bool) error {
	total := int32(0)
	for _, variant := range variants {
		if variant.ID == replacementID || !variant.IsActive {
			continue
		}
		total += variant.TrafficWeight
	}
	if replacementActive {
		total += replacementWeight
	}
	if total > 100 {
		return fmt.Errorf("%w: active variant traffic_weight total must not exceed 100", ErrWorkflowConfigurationInvalid)
	}
	return nil
}

func validateTrafficWeight(weight int32) error {
	if weight < 0 || weight > 100 {
		return fmt.Errorf("%w: traffic_weight must be between 0 and 100", ErrWorkflowConfigurationInvalid)
	}
	return nil
}

func normalizeConditions(raw json.RawMessage) ([]byte, string, error) {
	value, err := normalizeJSON(raw, []byte(`{}`))
	if err != nil {
		return nil, "", err
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, value); err != nil {
		return nil, "", fmt.Errorf("%w: invalid traffic_conditions", ErrWorkflowConfigurationInvalid)
	}
	sum := sha256.Sum256(compact.Bytes())
	return compact.Bytes(), hex.EncodeToString(sum[:]), nil
}

func normalizeExperimentStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "draft"
	}
	switch status {
	case "draft", "active", "paused", "completed", "cancelled", "archived":
		return status
	default:
		return ""
	}
}

func normalizeExperimentType(experimentType string) string {
	experimentType = strings.ToLower(strings.TrimSpace(experimentType))
	if experimentType == "" {
		return "split"
	}
	switch experimentType {
	case "split", "canary", "shadow", "rollout":
		return experimentType
	default:
		return ""
	}
}

func normalizeFallbackPolicy(policy string) string {
	policy = strings.ToLower(strings.TrimSpace(policy))
	if policy == "" {
		return "default_route"
	}
	return policy
}

func validateTimeRange(startedAt *time.Time, endedAt *time.Time) error {
	if startedAt != nil && endedAt != nil && !endedAt.After(*startedAt) {
		return fmt.Errorf("%w: ended_at must be after started_at", ErrWorkflowConfigurationInvalid)
	}
	return nil
}

func int32Value(value *int32, fallback int32) int32 {
	if value == nil {
		return fallback
	}
	return *value
}

func pgText(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func pgTimestamp(value *time.Time) pgtype.Timestamp {
	if value == nil || value.IsZero() {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: *value, Valid: true}
}
