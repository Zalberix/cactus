package message

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type runtimeRoutingStore interface {
	ListActiveExperimentScopesForRouting(ctx context.Context, arg db.ListActiveExperimentScopesForRoutingParams) ([]db.ListActiveExperimentScopesForRoutingRow, error)
	ListActiveWorkflowExperimentVariantsByScopeID(ctx context.Context, workflowExperimentScopeID int32) ([]db.WorkflowExperimentVariant, error)
	CountExperimentVariantRunsSince(ctx context.Context, arg db.CountExperimentVariantRunsSinceParams) ([]db.CountExperimentVariantRunsSinceRow, error)
	ListActiveRoutingCompatibilitiesByInputSchemaID(ctx context.Context, workflowInputSchemaID int32) ([]db.ListActiveRoutingCompatibilitiesByInputSchemaIDRow, error)
	GetDefaultRouteForInputSchema(ctx context.Context, workflowInputSchemaID int32) (db.WorkflowVersionInputSchemaCompatibility, error)
	GetActiveWorkflowVersionInputSchemaCompatibilityByPair(ctx context.Context, arg db.GetActiveWorkflowVersionInputSchemaCompatibilityByPairParams) (db.WorkflowVersionInputSchemaCompatibility, error)
	GetWorkflowInputMapperByID(ctx context.Context, id int32) (db.WorkflowInputMapper, error)
}

type runtimeRouteSelection struct {
	WorkflowVersionID           int32
	InputSchemaCompatibilityID  pgtype.Int4
	WorkflowInputMapperID       pgtype.Int4
	WorkflowExperimentID        pgtype.Int4
	WorkflowExperimentScopeID   pgtype.Int4
	WorkflowExperimentVariantID pgtype.Int4
	CompatibilityType           string
	DefaultValues               []byte
	SelectionReason             string
	VersionInputData            []byte
	RoutingDecision             []byte
}

const (
	selectionReasonStandard   = "standard"
	selectionReasonCanary     = "canary"
	selectionReasonExperiment = "experiment"
	selectionReasonRollout    = "rollout"
	selectionReasonFallback   = "fallback"
)

func canonicalSelectionReason(reason string) string {
	switch reason {
	case selectionReasonStandard, selectionReasonCanary, selectionReasonExperiment, selectionReasonRollout, selectionReasonFallback:
		return reason
	}

	switch reason {
	case "default_route":
		return selectionReasonStandard
	case "experiment_variant":
		return selectionReasonExperiment
	case "experiment_no_variant":
		return selectionReasonFallback
	case "experiment_scope_traffic_excluded":
		return selectionReasonFallback
	}

	if strings.HasPrefix(reason, "experiment_scope_traffic_excluded_") {
		return selectionReasonFallback
	}

	return selectionReasonStandard
}

func (s *Service) selectRuntimeRoute(ctx context.Context, workflowID int32, inputSchemaID int32, experimentID *int32, payloadJSON []byte, idempotencyKey string) (runtimeRouteSelection, error) {
	store, ok := s.store.(runtimeRoutingStore)
	if !ok {
		return runtimeRouteSelection{}, errors.New("runtime routing storage is unsupported")
	}

	payload, _ := decodeJSONObject(payloadJSON)
	seed := routeSelectionSeed(payload, idempotencyKey, newRandomRouteSeed)

	if selected, ok, err := s.selectExperimentRuntimeRoute(ctx, store, workflowID, inputSchemaID, experimentID, payload, seed); err != nil || ok {
		if err != nil {
			return runtimeRouteSelection{}, err
		}
		return s.withVersionInputData(ctx, store, selected, payloadJSON)
	}
	selected, err := s.selectDefaultRuntimeRoute(ctx, store, inputSchemaID, selectionReasonStandard)
	if err != nil {
		return runtimeRouteSelection{}, err
	}
	return s.withVersionInputData(ctx, store, selected, payloadJSON)
}

func (s *Service) selectExperimentRuntimeRoute(
	ctx context.Context,
	store runtimeRoutingStore,
	workflowID int32,
	inputSchemaID int32,
	experimentID *int32,
	payload map[string]any,
	seed string,
) (runtimeRouteSelection, bool, error) {
	var experimentIDParam pgtype.Int4
	if experimentID != nil && *experimentID > 0 {
		experimentIDParam = pgtype.Int4{Int32: *experimentID, Valid: true}
	}

	scopes, err := store.ListActiveExperimentScopesForRouting(ctx, db.ListActiveExperimentScopesForRoutingParams{
		WorkflowID:            workflowID,
		WorkflowInputSchemaID: inputSchemaID,
		ExperimentID:          experimentIDParam,
	})
	if err != nil {
		return runtimeRouteSelection{}, false, fmt.Errorf("list active experiment scopes: %w", err)
	}
	if experimentIDParam.Valid && len(scopes) == 0 {
		return runtimeRouteSelection{}, false, fmt.Errorf("EXPERIMENT_NOT_ROUTABLE: experiment %d is not active for workflow %d and input schema %d", experimentIDParam.Int32, workflowID, inputSchemaID)
	}

	for _, scope := range scopes {
		if !routeConditionsMatch(scope.TrafficConditions, payload) {
			continue
		}
		if scope.TrafficPercent < 100 {
			scopeBucket := bucketPercent(seed + ":scope:" + strconv.Itoa(int(scope.WorkflowExperimentScopeID)))
			if scopeBucket >= scope.TrafficPercent {
				selected, err := s.selectExperimentFallbackRuntimeRoute(ctx, store, inputSchemaID, scope, "experiment_scope_traffic_excluded")
				if err != nil {
					return runtimeRouteSelection{}, false, err
				}
				return selected, true, nil
			}
		}

		variants, err := store.ListActiveWorkflowExperimentVariantsByScopeID(ctx, scope.WorkflowExperimentScopeID)
		if err != nil {
			return runtimeRouteSelection{}, false, fmt.Errorf("list active experiment variants: %w", err)
		}
		variantCounts, err := countExperimentVariantRunsSinceTrafficChange(ctx, store, scope)
		if err != nil {
			return runtimeRouteSelection{}, false, err
		}
		variant, ok := selectMostUnderrepresentedVariant(variants, variantCounts)
		if !ok {
			selected, err := s.selectExperimentFallbackRuntimeRoute(ctx, store, inputSchemaID, scope, "experiment_no_variant")
			if err != nil {
				return runtimeRouteSelection{}, false, err
			}
			return selected, true, nil
		}
		selected, err := s.selectionForVersion(ctx, store, inputSchemaID, variant.WorkflowVersionID, "experiment_variant")
		if err != nil {
			return runtimeRouteSelection{}, false, err
		}
		selected.WorkflowExperimentID = routePgInt4(scope.WorkflowExperimentID)
		selected.WorkflowExperimentScopeID = routePgInt4(scope.WorkflowExperimentScopeID)
		selected.WorkflowExperimentVariantID = routePgInt4(variant.ID)
		selected.RoutingDecision = routingDecisionJSON(map[string]any{
			"reason":                         selected.SelectionReason,
			"workflow_experiment_id":         scope.WorkflowExperimentID,
			"workflow_experiment_scope_id":   scope.WorkflowExperimentScopeID,
			"workflow_experiment_variant_id": variant.ID,
			"workflow_version_id":            selected.WorkflowVersionID,
			"traffic_weight":                 variant.TrafficWeight,
		})
		return selected, true, nil
	}

	if experimentIDParam.Valid {
		return runtimeRouteSelection{}, false, fmt.Errorf("EXPERIMENT_NOT_ROUTABLE: experiment %d has no matching active scope for workflow %d and input schema %d", experimentIDParam.Int32, workflowID, inputSchemaID)
	}

	return runtimeRouteSelection{}, false, nil
}

func (s *Service) selectExperimentFallbackRuntimeRoute(
	ctx context.Context,
	store runtimeRoutingStore,
	inputSchemaID int32,
	scope db.ListActiveExperimentScopesForRoutingRow,
	reason string,
) (runtimeRouteSelection, error) {
	switch scope.FallbackPolicy {
	case "fallback_version":
		if !scope.FallbackWorkflowVersionID.Valid {
			return runtimeRouteSelection{}, fmt.Errorf("experiment scope %d requires fallback_workflow_version_id", scope.WorkflowExperimentScopeID)
		}
		selected, err := s.selectionForVersion(ctx, store, inputSchemaID, scope.FallbackWorkflowVersionID.Int32, reason+"_fallback_version")
		if err != nil {
			return runtimeRouteSelection{}, err
		}
		selected.WorkflowExperimentID = routePgInt4(scope.WorkflowExperimentID)
		selected.WorkflowExperimentScopeID = routePgInt4(scope.WorkflowExperimentScopeID)
		selected.RoutingDecision = routingDecisionJSON(map[string]any{
			"reason":                       selected.SelectionReason,
			"workflow_experiment_id":       scope.WorkflowExperimentID,
			"workflow_experiment_scope_id": scope.WorkflowExperimentScopeID,
			"workflow_version_id":          selected.WorkflowVersionID,
			"fallback_policy":              scope.FallbackPolicy,
		})
		return selected, nil
	case "default_route":
		selected, err := s.selectDefaultRuntimeRoute(ctx, store, inputSchemaID, reason+"_default_route")
		if err != nil {
			return runtimeRouteSelection{}, err
		}
		selected.WorkflowExperimentID = routePgInt4(scope.WorkflowExperimentID)
		selected.WorkflowExperimentScopeID = routePgInt4(scope.WorkflowExperimentScopeID)
		selected.RoutingDecision = routingDecisionJSON(map[string]any{
			"reason":                        selected.SelectionReason,
			"workflow_experiment_id":        scope.WorkflowExperimentID,
			"workflow_experiment_scope_id":  scope.WorkflowExperimentScopeID,
			"workflow_version_id":           selected.WorkflowVersionID,
			"input_schema_compatibility_id": selected.InputSchemaCompatibilityID.Int32,
			"fallback_policy":               scope.FallbackPolicy,
		})
		return selected, nil
	case "error", "":
		return runtimeRouteSelection{}, fmt.Errorf("experiment scope %d did not select a variant and fallback policy is error", scope.WorkflowExperimentScopeID)
	default:
		return runtimeRouteSelection{}, fmt.Errorf("experiment scope %d has unsupported fallback policy %q", scope.WorkflowExperimentScopeID, scope.FallbackPolicy)
	}
}

func (s *Service) selectDefaultRuntimeRoute(ctx context.Context, store runtimeRoutingStore, inputSchemaID int32, reason string) (runtimeRouteSelection, error) {
	route, err := store.GetDefaultRouteForInputSchema(ctx, inputSchemaID)
	if err != nil {
		active, listErr := store.ListActiveRoutingCompatibilitiesByInputSchemaID(ctx, inputSchemaID)
		if listErr != nil {
			return runtimeRouteSelection{}, fmt.Errorf("list active routing compatibilities: %w", listErr)
		}
		if len(active) == 0 {
			return runtimeRouteSelection{}, fmt.Errorf("no active workflow route for input schema %d: %w", inputSchemaID, err)
		}
		return runtimeRouteSelection{
			WorkflowVersionID:          active[0].WorkflowVersionID,
			InputSchemaCompatibilityID: routePgInt4(active[0].ID),
			WorkflowInputMapperID:      active[0].WorkflowInputMapperID,
			CompatibilityType:          active[0].CompatibilityType,
			DefaultValues:              active[0].DefaultValues,
			SelectionReason:            canonicalSelectionReason(reason),
			RoutingDecision: routingDecisionJSON(map[string]any{
				"reason":                        reason,
				"workflow_version_id":           active[0].WorkflowVersionID,
				"input_schema_compatibility_id": active[0].ID,
				"default_route":                 active[0].IsDefaultRoute,
			}),
		}, nil
	}
	return runtimeRouteSelection{
		WorkflowVersionID:          route.WorkflowVersionID,
		InputSchemaCompatibilityID: routePgInt4(route.ID),
		WorkflowInputMapperID:      route.WorkflowInputMapperID,
		CompatibilityType:          route.CompatibilityType,
		DefaultValues:              route.DefaultValues,
		SelectionReason:            canonicalSelectionReason(reason),
		RoutingDecision: routingDecisionJSON(map[string]any{
			"reason":                        reason,
			"workflow_version_id":           route.WorkflowVersionID,
			"input_schema_compatibility_id": route.ID,
			"default_route":                 route.IsDefaultRoute,
		}),
	}, nil
}

func (s *Service) selectionForVersion(ctx context.Context, store runtimeRoutingStore, inputSchemaID int32, versionID int32, reason string) (runtimeRouteSelection, error) {
	compatibility, err := store.GetActiveWorkflowVersionInputSchemaCompatibilityByPair(ctx, db.GetActiveWorkflowVersionInputSchemaCompatibilityByPairParams{
		WorkflowVersionID:     versionID,
		WorkflowInputSchemaID: inputSchemaID,
	})
	if err != nil {
		return runtimeRouteSelection{}, fmt.Errorf("workflow version %d is not active-compatible with input schema %d: %w", versionID, inputSchemaID, err)
	}
	return runtimeRouteSelection{
		WorkflowVersionID:          versionID,
		InputSchemaCompatibilityID: routePgInt4(compatibility.ID),
		WorkflowInputMapperID:      compatibility.WorkflowInputMapperID,
		CompatibilityType:          compatibility.CompatibilityType,
		DefaultValues:              compatibility.DefaultValues,
		SelectionReason:            canonicalSelectionReason(reason),
		RoutingDecision: routingDecisionJSON(map[string]any{
			"reason":                        reason,
			"workflow_version_id":           versionID,
			"input_schema_compatibility_id": compatibility.ID,
		}),
	}, nil
}

func (s *Service) withVersionInputData(ctx context.Context, store runtimeRoutingStore, route runtimeRouteSelection, payloadJSON []byte) (runtimeRouteSelection, error) {
	versionInputData, err := s.buildVersionInputData(ctx, store, route, payloadJSON)
	if err != nil {
		return runtimeRouteSelection{}, err
	}
	route.VersionInputData = versionInputData
	return route, nil
}

func (s *Service) buildVersionInputData(ctx context.Context, store runtimeRoutingStore, route runtimeRouteSelection, payloadJSON []byte) ([]byte, error) {
	input, err := decodeJSONObject(payloadJSON)
	if err != nil {
		return nil, fmt.Errorf("decode message payload: %w", err)
	}

	versionInput := input
	if route.WorkflowInputMapperID.Valid {
		mapper, err := store.GetWorkflowInputMapperByID(ctx, route.WorkflowInputMapperID.Int32)
		if err != nil {
			return nil, fmt.Errorf("get workflow input mapper: %w", err)
		}
		if !mapper.IsActive {
			return nil, fmt.Errorf("workflow input mapper %d is inactive", mapper.ID)
		}
		versionInput, err = applyWorkflowInputMapper(mapper, input)
		if err != nil {
			return nil, err
		}
	}

	defaults, err := decodeOptionalJSONObject(route.DefaultValues)
	if err != nil {
		return nil, fmt.Errorf("decode compatibility default values: %w", err)
	}
	versionInput = mergeDefaultValues(defaults, versionInput)
	return json.Marshal(versionInput)
}

func applyWorkflowInputMapper(mapper db.WorkflowInputMapper, input map[string]any) (map[string]any, error) {
	mapperType := strings.ToLower(strings.TrimSpace(mapper.MapperType))
	if mapperType == "" || mapperType == "identity" || mapperType == "native" {
		return cloneMap(input), nil
	}

	rules, err := decodeOptionalJSONObject(mapper.Rules)
	if err != nil {
		return nil, fmt.Errorf("decode workflow input mapper rules: %w", err)
	}
	if len(rules) == 0 {
		return cloneMap(input), nil
	}

	output := map[string]any{}
	if copyAll, _ := rules["copy_all"].(bool); copyAll {
		output = cloneMap(input)
	}
	if defaults, ok := objectValue(rules["defaults"]); ok {
		output = mergeDefaultValues(defaults, output)
	}

	mappings, ok := mapperMappings(rules)
	if !ok {
		return output, nil
	}
	for targetPath, sourceSpec := range mappings {
		value, exists := mapperSourceValue(sourceSpec, input)
		if !exists {
			continue
		}
		setPathValue(output, targetPath, value)
	}
	return output, nil
}

func mapperMappings(rules map[string]any) (map[string]any, bool) {
	for _, key := range []string{"fields", "mapping", "mappings"} {
		if mappings, ok := objectValue(rules[key]); ok {
			return mappings, true
		}
	}
	return nil, false
}

func mapperSourceValue(sourceSpec any, input map[string]any) (any, bool) {
	switch value := sourceSpec.(type) {
	case string:
		sourcePath := strings.TrimPrefix(strings.TrimPrefix(value, "$."), "payload.")
		return getPathValue(input, sourcePath)
	case map[string]any:
		if literal, ok := value["literal"]; ok {
			return literal, true
		}
		if source, ok := value["source"].(string); ok {
			sourcePath := strings.TrimPrefix(strings.TrimPrefix(source, "$."), "payload.")
			return getPathValue(input, sourcePath)
		}
		if source, ok := value["path"].(string); ok {
			sourcePath := strings.TrimPrefix(strings.TrimPrefix(source, "$."), "payload.")
			return getPathValue(input, sourcePath)
		}
	}
	return nil, false
}

func decodeOptionalJSONObject(raw []byte) (map[string]any, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte(`null`)) {
		return map[string]any{}, nil
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if value == nil {
		return map[string]any{}, nil
	}
	return value, nil
}

func mergeDefaultValues(defaults map[string]any, values map[string]any) map[string]any {
	merged := cloneMap(defaults)
	for key, value := range values {
		merged[key] = value
	}
	return merged
}

func cloneMap(value map[string]any) map[string]any {
	clone := make(map[string]any, len(value))
	for key, item := range value {
		clone[key] = item
	}
	return clone
}

func objectValue(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	return object, ok
}

func getPathValue(input map[string]any, path string) (any, bool) {
	if path == "" {
		return nil, false
	}
	var current any = input
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func setPathValue(output map[string]any, path string, value any) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(path, "$."), "payload."), ".")
	if len(parts) == 0 || parts[0] == "" {
		return
	}
	current := output
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = value
}

func countExperimentVariantRunsSinceTrafficChange(
	ctx context.Context,
	store runtimeRoutingStore,
	scope db.ListActiveExperimentScopesForRoutingRow,
) (map[int32]int32, error) {
	rows, err := store.CountExperimentVariantRunsSince(ctx, db.CountExperimentVariantRunsSinceParams{
		WorkflowExperimentScopeID: routePgInt4(scope.WorkflowExperimentScopeID),
		CreatedAt:                 scope.TrafficChangedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("count experiment variant runs since traffic change: %w", err)
	}

	counts := make(map[int32]int32, len(rows))
	for _, row := range rows {
		counts[row.WorkflowExperimentVariantID] = row.RunCount
	}
	return counts, nil
}

func selectMostUnderrepresentedVariant(variants []db.WorkflowExperimentVariant, counts map[int32]int32) (db.WorkflowExperimentVariant, bool) {
	total := int32(0)
	for _, variant := range variants {
		if variant.IsActive && variant.TrafficWeight > 0 {
			total += variant.TrafficWeight
		}
	}
	if total <= 0 {
		for _, variant := range variants {
			if variant.IsActive && variant.IsControlGroup {
				return variant, true
			}
		}
		return db.WorkflowExperimentVariant{}, false
	}

	totalRuns := int64(0)
	for _, variant := range variants {
		if !variant.IsActive || variant.TrafficWeight <= 0 {
			continue
		}
		totalRuns += int64(counts[variant.ID])
	}

	var selected db.WorkflowExperimentVariant
	bestDeficit := int64(0)
	hasSelected := false
	for _, variant := range variants {
		if !variant.IsActive || variant.TrafficWeight <= 0 {
			continue
		}
		actual := int64(counts[variant.ID])
		deficit := (totalRuns+1)*int64(variant.TrafficWeight) - actual*int64(total)
		if !hasSelected || deficit > bestDeficit {
			selected = variant
			bestDeficit = deficit
			hasSelected = true
		}
	}
	return selected, hasSelected
}

func routeConditionsMatch(raw []byte, payload map[string]any) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte(`{}`)) {
		return true
	}
	var conditions map[string]any
	if err := json.Unmarshal(raw, &conditions); err != nil {
		return false
	}
	if len(conditions) == 0 {
		return true
	}
	if fields, ok := conditions["fields"].(map[string]any); ok {
		return allFieldConditionsMatch(fields, payload)
	}
	return allFieldConditionsMatch(conditions, payload)
}

func allFieldConditionsMatch(conditions map[string]any, payload map[string]any) bool {
	for field, expected := range conditions {
		if field == "all" || field == "any" {
			continue
		}
		if !fieldConditionMatch(payload[field], expected) {
			return false
		}
	}
	return true
}

func fieldConditionMatch(actual any, expected any) bool {
	if operator, ok := expected.(map[string]any); ok {
		for op, value := range operator {
			switch op {
			case "eq":
				if !jsonValueEqual(actual, value) {
					return false
				}
			case "neq":
				if jsonValueEqual(actual, value) {
					return false
				}
			case "exists":
				want, _ := value.(bool)
				if (actual != nil) != want {
					return false
				}
			case "in":
				values, ok := value.([]any)
				if !ok {
					return false
				}
				matched := false
				for _, candidate := range values {
					if jsonValueEqual(actual, candidate) {
						matched = true
						break
					}
				}
				if !matched {
					return false
				}
			}
		}
		return true
	}
	return jsonValueEqual(actual, expected)
}

func jsonValueEqual(a any, b any) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return bytes.Equal(aJSON, bJSON)
}

func decodeJSONObject(raw []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{}, err
	}
	return payload, nil
}

func canonicalRouteSeed(payload map[string]any) string {
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]any, len(keys))
	for _, key := range keys {
		ordered[key] = payload[key]
	}
	raw, _ := json.Marshal(ordered)
	return string(raw)
}

func routeSelectionSeed(payload map[string]any, idempotencyKey string, randomSeed func() string) string {
	seed := strings.TrimSpace(idempotencyKey)
	if seed != "" {
		return seed
	}

	if randomSeed != nil {
		if seed = strings.TrimSpace(randomSeed()); seed != "" {
			return seed
		}
	}

	return canonicalRouteSeed(payload)
}

func newRandomRouteSeed() string {
	var seed [16]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(seed[:])
}

func bucketPercent(seed string) int32 {
	hash := sha256.Sum256([]byte(seed))
	value := binary.BigEndian.Uint64(hash[:8])
	return int32(value % 100)
}

func routingDecisionJSON(value map[string]any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func routePgInt4(value int32) pgtype.Int4 {
	if value <= 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: value, Valid: true}
}
