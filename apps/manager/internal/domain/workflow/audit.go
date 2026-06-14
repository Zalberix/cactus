package workflow

import (
	"context"
	"encoding/json"

	db "github.com/zalberix/cactus/libs/storage/db"
)

type workflowConfigurationAuditQueries interface {
	CreateWorkflowConfigurationAuditLog(ctx context.Context, arg db.CreateWorkflowConfigurationAuditLogParams) (db.WorkflowConfigurationAuditLog, error)
}

func (s *Service) auditWorkflowConfiguration(ctx context.Context, entityType string, entityID int32, action string, actorUserID int32, beforeValue any, afterValue any) error {
	q, ok := s.store.(workflowConfigurationAuditQueries)
	if !ok {
		return nil
	}
	return auditWorkflowConfigurationWithQueries(ctx, q, entityType, entityID, action, actorUserID, beforeValue, afterValue)
}

func auditWorkflowConfigurationWithQueries(ctx context.Context, q workflowConfigurationAuditQueries, entityType string, entityID int32, action string, actorUserID int32, beforeValue any, afterValue any) error {
	if q == nil {
		return nil
	}
	beforeJSON, err := auditJSON(beforeValue)
	if err != nil {
		return err
	}
	afterJSON, err := auditJSON(afterValue)
	if err != nil {
		return err
	}
	_, err = q.CreateWorkflowConfigurationAuditLog(ctx, db.CreateWorkflowConfigurationAuditLogParams{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		ActorUserID: pgInt4(actorUserID),
		BeforeValue: beforeJSON,
		AfterValue:  afterJSON,
	})
	return err
}

func auditJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte(`null`), nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
