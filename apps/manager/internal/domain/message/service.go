package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/client"

	"github.com/zalberix/cactus/apps/manager/internal/http/response"
	temporaltypes "github.com/zalberix/cactus/apps/manager/internal/temporal"
	db "github.com/zalberix/cactus/libs/storage/db"
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
// 4. Валидировать payload по input_schema
// 5. Создать message в БД
// 6. Сформировать DAGInput из steps + deps активной версии
// 7. Создать workflow_run в БД
// 8. Запустить Temporal workflow
// 9. Обновить workflow_run с temporal_workflow_id
// 10. Вернуть response
func (s *Service) SendMessage(ctx context.Context, req SendMessageRequest, publicToken string) (*SendMessageResponse, []response.ErrorDetail, error) {
	if strings.TrimSpace(req.Process) != "" {
		ref, err := parseProcessRef(req.Process)
		if err != nil {
			return nil, nil, err
		}
		if req.WorkflowID != 0 && req.WorkflowID != ref.WorkflowID {
			return nil, nil, fmt.Errorf("PROCESS_WORKFLOW_MISMATCH")
		}
		req.WorkflowID = ref.WorkflowID
	}
	if req.WorkflowID <= 0 {
		return nil, nil, fmt.Errorf("PROCESS_REQUIRED: process is required")
	}

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

	resolvedSchema, err := s.resolveInputSchema(ctx, req, wf)
	if err != nil {
		return nil, nil, err
	}

	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal value: %w", err)
	}
	if existing, found, err := s.findExistingIdempotentMessage(ctx, req, valueJSON); err != nil {
		return nil, nil, err
	} else if found {
		existing.WorkflowInputSchemaCode = resolvedSchema.Code
		existing.WorkflowInputSchemaVersionNumber = pgInt4Ptr(resolvedSchema.VersionNumber)
		return existing, nil, nil
	}

	// 3. Найти маршрут выполнения по input schema, compatibility и active experiments.
	route, err := s.selectRuntimeRoute(ctx, req.WorkflowID, resolvedSchema.ID.Int32, req.Experimental, valueJSON, req.IdempotencyKey)
	if err != nil {
		return nil, nil, fmt.Errorf("select workflow route: %w", err)
	}

	// 4. Валидировать payload по JSON Schema
	if validationErrors := ValidatePayload(resolvedSchema.SchemaJSON, req.Value); len(validationErrors) > 0 {
		return nil, validationErrors, nil
	}

	metadataJSON := []byte(`{}`)
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal metadata: %w", err)
		}
	}
	var overriddenPriority pgtype.Int4
	if req.OverriddenPriority != nil {
		overriddenPriority = pgtype.Int4{Int32: *req.OverriddenPriority, Valid: true}
	}
	idempotencyKey := pgtype.Text{String: req.IdempotencyKey, Valid: req.IdempotencyKey != ""}
	msg, err := s.store.CreateNewMessage(ctx, db.CreateNewMessageParams{
		WorkflowID:            req.WorkflowID,
		WorkflowInputSchemaID: resolvedSchema.ID,
		ExternalMessageID:     pgtype.Text{String: req.ExternalID, Valid: req.ExternalID != ""},
		IdempotencyKey:        idempotencyKey,
		OverriddenPriority:    overriddenPriority,
		Value:                 valueJSON,
		Column7:               metadataJSON,
		Status:                "created",
		ErrorMessage:          pgtype.Text{},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create message: %w", err)
	}

	// 6. Сформировать DAGInput
	dagInput, err := s.buildDAGInput(ctx, route.WorkflowVersionID, msg.ID, route.VersionInputData)
	if err != nil {
		return nil, nil, fmt.Errorf("build DAG input: %w", err)
	}

	// 7. Создать workflow_run (ID будет использован как WorkflowRunID в DAGInput)
	run, err := s.store.CreateWorkflowRun(ctx, db.CreateWorkflowRunParams{
		WorkflowVersionID:           route.WorkflowVersionID,
		MessageID:                   msg.ID,
		TemporalWorkflowID:          pgtype.Text{Valid: false},
		Status:                      temporaltypes.RunStatusRunning,
		WorkflowExperimentID:        route.WorkflowExperimentID,
		WorkflowExperimentScopeID:   route.WorkflowExperimentScopeID,
		WorkflowExperimentVariantID: route.WorkflowExperimentVariantID,
		InputSchemaCompatibilityID:  route.InputSchemaCompatibilityID,
		Column9:                     route.SelectionReason,
		Column10:                    route.VersionInputData,
		RoutingDecision:             route.RoutingDecision,
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
		MessageID:                        msg.ID,
		WorkflowRunID:                    run.ID,
		WorkflowInputSchemaID:            pgInt4Ptr(resolvedSchema.ID),
		WorkflowInputSchemaCode:          resolvedSchema.Code,
		WorkflowInputSchemaVersionNumber: pgInt4Ptr(resolvedSchema.VersionNumber),
		WorkflowVersionID:                int32Ptr(route.WorkflowVersionID),
		InputSchemaCompatibilityID:       pgInt4Ptr(route.InputSchemaCompatibilityID),
		WorkflowExperimentID:             pgInt4Ptr(route.WorkflowExperimentID),
		WorkflowExperimentScopeID:        pgInt4Ptr(route.WorkflowExperimentScopeID),
		WorkflowExperimentVariantID:      pgInt4Ptr(route.WorkflowExperimentVariantID),
		SelectionReason:                  route.SelectionReason,
		Status:                           "running",
	}, nil, nil
}

// buildDAGInput собирает DAGInput из steps и deps активной версии.
func (s *Service) buildDAGInput(ctx context.Context, versionID, messageID int32, messageValue []byte) (*temporaltypes.DAGInput, error) {
	// Загружаем шаги
	steps, err := s.store.ListEnrichedStepsByVersionID(ctx, versionID)
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
		if len(step.ControlSettings) > 0 {
			sd.ControlSettings = step.ControlSettings
		}
		if step.WorkTypeID.Valid {
			sd.WorkTypeID = step.WorkTypeID.Int32
		}
		if step.OrganizationID.Valid {
			sd.OrganizationID = step.OrganizationID.Int32
		}
		if step.WorkerSettingsRevisionID.Valid {
			sd.WorkerSettingsRevisionID = step.WorkerSettingsRevisionID.Int32
		}
		if step.WorkerSettingsSchemaID.Valid {
			sd.WorkerSettingsSchemaID = step.WorkerSettingsSchemaID.Int32
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
		Limit:          int32(perPage), // #nosec G115 -- perPage is bounded above.
		Offset:         int32(offset),  // #nosec G115 -- offset is derived from pagination input.
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}

	items := make([]ListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ListItem{
			ID:                               r.ID,
			WorkflowID:                       r.WorkflowID,
			WorkflowName:                     r.WorkflowName,
			WorkflowInputSchemaID:            pgInt4Ptr(r.WorkflowInputSchemaID),
			WorkflowInputSchemaCode:          pgTextPtr(r.WorkflowInputSchemaCode),
			WorkflowInputSchemaVersionNumber: pgInt4Ptr(r.WorkflowInputSchemaVersionNumber),
			WorkflowVersionID:                pgInt4Ptr(r.WorkflowVersionID),
			WorkflowVersionNumber:            pgInt4Ptr(r.WorkflowVersionNumber),
			WorkflowVersionName:              pgTextPtr(r.WorkflowVersionName),
			Status:                           r.Status,
			CreatedAt:                        r.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:                        r.UpdatedAt.Time.Format(time.RFC3339),
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
		now := time.Now().UTC()
		for _, step := range steps {
			dto := StepStatusDTO{
				ID:         step.ID,
				StepID:     step.WorkflowStepID,
				StepType:   step.StepType,
				Status:     step.Status,
				DurationMs: stepDurationMs(step.StartedAt, step.CompletedAt, now),
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

func (s *Service) GetMessageDetail(ctx context.Context, messageID int32) (*DetailResponse, error) {
	row, err := s.store.GetMessageDetailByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message detail not found: %w", err)
	}

	resp := buildMessageDetailBase(row)
	if !row.WorkflowRunID.Valid {
		return resp, nil
	}

	versionID := row.WorkflowVersionID.Int32
	steps, err := s.store.ListEnrichedStepsByVersionID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("list run graph steps: %w", err)
	}
	deps, err := s.store.ListDependenciesByVersionID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("list run graph dependencies: %w", err)
	}
	runSteps, err := s.store.ListWorkflowRunStepDetailsByRunID(ctx, row.WorkflowRunID.Int32)
	if err != nil {
		return nil, fmt.Errorf("list run step details: %w", err)
	}

	resp.Graph = buildMessageGraph(versionID, steps, deps)
	resp.RunSteps = buildRunStepDetails(runSteps)
	return resp, nil
}

func buildMessageDetailBase(row db.GetMessageDetailByIDRow) *DetailResponse {
	messageValue := jsonObjectFromBytes(row.MessageValue)
	if messageValue == nil {
		messageValue = map[string]any{}
	}

	resp := &DetailResponse{
		MessageID:             row.ID,
		WorkflowID:            row.WorkflowID,
		WorkflowName:          row.WorkflowName,
		WorkflowVersionID:     pgInt4Ptr(row.WorkflowVersionID),
		WorkflowVersionNumber: pgInt4Ptr(row.WorkflowVersionNumber),
		WorkflowVersionName:   pgTextPtr(row.WorkflowVersionName),
		MessageStatus:         row.MessageStatus,
		MessageValue:          messageValue,
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
		Graph: GraphDTO{
			Steps:        []GraphStepDTO{},
			Dependencies: []GraphDependencyDTO{},
		},
		RunSteps: []StepRunDetailDTO{},
	}

	if !row.WorkflowRunID.Valid {
		return resp
	}

	run := &WorkflowRunStatus{
		ID:     row.WorkflowRunID.Int32,
		Status: row.WorkflowStatus.String,
	}
	if row.RunStartedAt.Valid {
		t := row.RunStartedAt.Time
		run.StartedAt = &t
	}
	if row.RunCompletedAt.Valid {
		t := row.RunCompletedAt.Time
		run.CompletedAt = &t
	}
	if row.RunErrorMessage.Valid {
		run.ErrorMessage = &row.RunErrorMessage.String
	}
	resp.WorkflowRun = run
	return resp
}

func buildMessageGraph(versionID int32, steps []db.ListEnrichedStepsByVersionIDRow, deps []db.WorkflowStepDependency) GraphDTO {
	graph := GraphDTO{
		VersionID:    versionID,
		Steps:        make([]GraphStepDTO, 0, len(steps)),
		Dependencies: make([]GraphDependencyDTO, 0, len(deps)),
	}

	for _, step := range steps {
		dto := GraphStepDTO{
			ID:             step.ID,
			StepType:       step.StepType,
			Name:           step.Name,
			WorkTypeMeta:   jsonObjectFromBytes(step.WorkTypeMeta),
			InputMapping:   mappingFromRawJSON(step.InputMapping),
			CanvasPosition: jsonObjectFromBytes(step.CanvasPosition),
			InputSchema:    jsonObjectFromBytes(step.InputSchema),
			OutputSchema:   jsonObjectFromBytes(step.OutputSchema),
		}
		if step.ControlKind.Valid {
			dto.ControlKind = &step.ControlKind.String
		}
		if step.WorkTypeID.Valid {
			dto.WorkTypeID = &step.WorkTypeID.Int32
		}
		if step.WorkTypeName.Valid {
			dto.WorkTypeName = &step.WorkTypeName.String
		}
		if step.WorkTypeCode.Valid {
			dto.WorkTypeCode = &step.WorkTypeCode.String
		}
		graph.Steps = append(graph.Steps, dto)
	}

	for _, dep := range deps {
		outcome := ""
		if dep.Outcome.Valid {
			outcome = dep.Outcome.String
		}
		graph.Dependencies = append(graph.Dependencies, GraphDependencyDTO{
			StepID:          dep.StepID,
			DependsOnStepID: dep.DependsOnStepID,
			Outcome:         outcome,
			OutputIndex:     dep.OutputIndex,
		})
	}

	return graph
}

func buildRunStepDetails(runSteps []db.WorkflowRunStep) []StepRunDetailDTO {
	result := make([]StepRunDetailDTO, 0, len(runSteps))
	now := time.Now().UTC()
	for _, step := range runSteps {
		dto := StepRunDetailDTO{
			ID:         step.ID,
			StepID:     step.WorkflowStepID,
			Status:     step.Status,
			InputData:  jsonObjectFromBytes(step.InputData),
			OutputData: jsonObjectFromBytes(step.OutputData),
			DurationMs: stepDurationMs(step.StartedAt, step.CompletedAt, now),
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
		result = append(result, dto)
	}
	return result
}

func stepDurationMs(startedAt, completedAt pgtype.Timestamp, now time.Time) *int64 {
	if !startedAt.Valid {
		return nil
	}
	end := now
	if completedAt.Valid {
		end = completedAt.Time
	}
	ms := end.Sub(startedAt.Time).Milliseconds()
	if ms < 0 {
		ms = 0
	}
	return &ms
}

func pgInt4Ptr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func pgTextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func jsonObjectFromBytes(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return value
}

func mappingFromRawJSON(raw []byte) []MappingDTO {
	if len(raw) == 0 {
		return []MappingDTO{}
	}
	var mapping []MappingDTO
	if err := json.Unmarshal(raw, &mapping); err != nil {
		return []MappingDTO{}
	}
	return mapping
}
