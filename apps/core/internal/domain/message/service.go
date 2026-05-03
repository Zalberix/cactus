package message

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/client"

	"github.com/zalberix/cactus/apps/core/internal/http/response"
	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// Service — бизнес-логика message domain.
type Service struct {
	store          Storage
	temporalClient client.Client
}

// NewService создаёт новый Service с заданным хранилищем и Temporal client.
func NewService(store Storage, tc client.Client) *Service {
	return &Service{
		store:          store,
		temporalClient: tc,
	}
}

// SendMessage — основная бизнес-логика отправки сообщения.
//
// Алгоритм:
// 1. Загрузить workflow по ID
// 2. Проверить доступ (system token → CheckWorkflowAccess)
// 3. Найти единственную активную версию
// 4. Валидировать payload по input_validation
// 5. Создать message в БД
// 6. Сформировать DAGInput из steps + deps активной версии
// 7. Создать workflow_run в БД
// 8. Запустить Temporal workflow
// 9. Обновить workflow_run с temporal_workflow_id
// 10. Вернуть response
func (s *Service) SendMessage(ctx context.Context, req SendMessageRequest, publicToken string) (*SendMessageResponse, []response.ErrorDetail, error) {
	// 1. Загрузить workflow
	wf, err := s.store.GetWorkflowByID(ctx, req.WorkflowID)
	if err != nil {
		return nil, nil, fmt.Errorf("workflow not found: %w", err)
	}

	// 2. Проверить доступ по system token (если передан)
	if publicToken != "" {
		hasAccess, err := s.store.CheckWorkflowAccess(ctx, db.CheckWorkflowAccessParams{
			WorkflowID:  req.WorkflowID,
			PublicToken: publicToken,
		})
		if err != nil || !hasAccess {
			return nil, nil, fmt.Errorf("access denied to workflow %d", req.WorkflowID)
		}
	}

	// 3. Найти активную версию
	activeVersions, err := s.store.ListActiveWorkflowVersions(ctx, req.WorkflowID)
	if err != nil {
		return nil, nil, fmt.Errorf("list active versions: %w", err)
	}
	if len(activeVersions) == 0 {
		return nil, nil, fmt.Errorf("no active version for workflow %d", req.WorkflowID)
	}
	activeVersion := activeVersions[0]

	// 4. Валидировать payload по JSON Schema
	if validationErrors := ValidatePayload(wf.InputValidation, req.Value); len(validationErrors) > 0 {
		return nil, validationErrors, nil
	}

	// 5. Создать message
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal value: %w", err)
	}
	msg, err := s.store.CreateNewMessage(ctx, db.CreateNewMessageParams{
		WorkflowID:        req.WorkflowID,
		ExternalMessageID: pgtype.Text{String: req.ExternalID, Valid: req.ExternalID != ""},
		Value:             valueJSON,
		Status:            "created",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create message: %w", err)
	}

	// 6. Сформировать DAGInput
	dagInput, err := s.buildDAGInput(ctx, activeVersion.ID, msg.ID, valueJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("build DAG input: %w", err)
	}

	// 7. Создать workflow_run (ID будет использован как WorkflowRunID в DAGInput)
	run, err := s.store.CreateWorkflowRun(ctx, db.CreateWorkflowRunParams{
		WorkflowVersionID:  activeVersion.ID,
		MessageID:          msg.ID,
		TemporalWorkflowID: pgtype.Text{Valid: false},
		Status:             temporaltypes.RunStatusRunning,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create workflow_run: %w", err)
	}

	dagInput.WorkflowRunID = run.ID

	// 8. Запустить Temporal workflow (per D-13: синхронный StartWorkflow)
	// Используем строковое имя workflow для избежания circular import с temporal/workflow
	workflowID := fmt.Sprintf("dag-%d-%d", req.WorkflowID, msg.ID)
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: temporaltypes.TaskQueueName,
	}

	_, err = s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "DAGExecutorWorkflow", dagInput)
	if err != nil {
		return nil, nil, fmt.Errorf("start temporal workflow: %w", err)
	}

	// 9. Обновить workflow_run (ставит status=running и started_at)
	_, _ = s.store.UpdateWorkflowRunStarted(ctx, run.ID)

	// 10. Ответ (per D-14)
	return &SendMessageResponse{
		MessageID:     msg.ID,
		WorkflowRunID: run.ID,
		Status:        "running",
	}, nil, nil
}

// buildDAGInput собирает DAGInput из steps и deps активной версии.
func (s *Service) buildDAGInput(ctx context.Context, versionID, messageID int32, messageValue []byte) (*temporaltypes.DAGInput, error) {
	// Загружаем шаги
	steps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("list steps: %w", err)
	}

	// Загружаем зависимости
	deps, err := s.store.ListDependenciesByVersionID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}

	// Конвертируем в temporal types
	stepDefs := make([]temporaltypes.StepDef, 0, len(steps))
	for _, step := range steps {
		sd := temporaltypes.StepDef{
			ID:       step.ID,
			StepType: step.StepType,
		}
		if step.ControlKind.Valid {
			sd.ControlKind = step.ControlKind.String
		}
		if step.WorkTypeID.Valid {
			sd.WorkTypeID = step.WorkTypeID.Int32
		}
		if step.WorkerSettingsRevisionID.Valid {
			sd.WorkerSettingsRevisionID = step.WorkerSettingsRevisionID.Int32
		}
		// Per D-05: Timeout should be populated from WorkerSettingsRevision.
		// In v1, we leave it as zero (default worker timeout will be used in activity).
		// Future: query WorkerSettingsRevision and set sd.Timeout from its timeout field.

		// Десериализуем input_mapping из JSONB
		if len(step.InputMapping) > 0 {
			var mapping []temporaltypes.MappingEntry
			if err := json.Unmarshal(step.InputMapping, &mapping); err != nil {
				return nil, fmt.Errorf("unmarshal input_mapping step %d: %w", step.ID, err)
			}
			sd.InputMapping = mapping
		}
		stepDefs = append(stepDefs, sd)
	}

	depDefs := make([]temporaltypes.DepDef, 0, len(deps))
	for _, d := range deps {
		dd := temporaltypes.DepDef{
			StepID:          d.StepID,
			DependsOnStepID: d.DependsOnStepID,
		}
		if d.Outcome.Valid {
			dd.Outcome = d.Outcome.String
		}
		depDefs = append(depDefs, dd)
	}

	return &temporaltypes.DAGInput{
		MessageID:    messageID,
		MessageValue: messageValue,
		Steps:        stepDefs,
		Deps:         depDefs,
	}, nil
}

// ListMessages возвращает пагинированный список сообщений организации (per UI-12).
func (s *Service) ListMessages(ctx context.Context, orgID int32, page, perPage int) ([]ListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Get total count
	total, err := s.store.CountMessagesByOrganizationID(ctx, pgtype.Int4{Int32: orgID, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}

	// Get page of messages
	rows, err := s.store.ListMessagesByOrganizationID(ctx, db.ListMessagesByOrganizationIDParams{
		OrganizationID: pgtype.Int4{Int32: orgID, Valid: true},
		Limit:          int64(perPage),
		Offset:         int64(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}

	items := make([]ListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ListItem{
			ID:           r.ID,
			WorkflowID:   r.WorkflowID,
			WorkflowName: r.WorkflowName,
			Status:       r.Status,
			CreatedAt:    r.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:    r.UpdatedAt.Time.Format(time.RFC3339),
		})
	}

	return items, total, nil
}

// GetMessageStatus returns message status with workflow run and step statuses (per EXEC-09, D-20, D-21).
func (s *Service) GetMessageStatus(ctx context.Context, messageID int32) (*StatusResponse, error) {
	// 1. Query message + workflow_run
	row, err := s.store.GetMessageStatusByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	resp := &StatusResponse{
		MessageID:     row.ID,
		MessageStatus: row.MessageStatus,
		CreatedAt:     row.CreatedAt.Time,
		Steps:         []StepStatusDTO{},
	}

	if !row.WorkflowRunID.Valid {
		return resp, nil
	}

	wrs := &WorkflowRunStatus{
		ID:     row.WorkflowRunID.Int32,
		Status: row.WorkflowStatus.String,
	}
	if row.RunStartedAt.Valid {
		t := row.RunStartedAt.Time
		wrs.StartedAt = &t
	}
	if row.RunCompletedAt.Valid {
		t := row.RunCompletedAt.Time
		wrs.CompletedAt = &t
	}
	if row.RunErrorMessage.Valid {
		wrs.ErrorMessage = &row.RunErrorMessage.String
	}
	resp.WorkflowRun = wrs

	if steps, err := s.store.ListWorkflowRunStepStatusesByRunID(ctx, row.WorkflowRunID.Int32); err == nil {
		for _, step := range steps {
			dto := StepStatusDTO{
				ID:       step.ID,
				StepID:   step.WorkflowStepID,
				StepType: step.StepType,
				Status:   step.Status,
			}
			if step.Outcome.Valid {
				dto.Outcome = &step.Outcome.String
			}
			if step.StartedAt.Valid {
				t := step.StartedAt.Time
				dto.StartedAt = &t
			}
			if step.CompletedAt.Valid {
				t := step.CompletedAt.Time
				dto.CompletedAt = &t
			}
			if step.ErrorMessage.Valid {
				dto.ErrorMessage = &step.ErrorMessage.String
			}
			resp.Steps = append(resp.Steps, dto)
		}
	}

	return resp, nil
}
