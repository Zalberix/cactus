package message

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"

	db "github.com/zalberix/cactus/libs/storage/db"
)

type runtimeRoutingTestStore struct {
	detailStore

	scopes        []db.ListActiveExperimentScopesForRoutingRow
	variants      map[int32][]db.WorkflowExperimentVariant
	counts        map[int32][]db.CountExperimentVariantRunsSinceRow
	defaultRoute  db.WorkflowVersionInputSchemaCompatibility
	pairRoutes    map[[2]int32]db.WorkflowVersionInputSchemaCompatibility
	mapper        db.WorkflowInputMapper
	routingParams db.ListActiveExperimentScopesForRoutingParams
}

func (s *runtimeRoutingTestStore) ListActiveExperimentScopesForRouting(_ context.Context, arg db.ListActiveExperimentScopesForRoutingParams) ([]db.ListActiveExperimentScopesForRoutingRow, error) {
	s.routingParams = arg
	return s.scopes, nil
}

func (s *runtimeRoutingTestStore) ListActiveWorkflowExperimentVariantsByScopeID(_ context.Context, workflowExperimentScopeID int32) ([]db.WorkflowExperimentVariant, error) {
	return s.variants[workflowExperimentScopeID], nil
}

func (s *runtimeRoutingTestStore) CountExperimentVariantRunsSince(_ context.Context, arg db.CountExperimentVariantRunsSinceParams) ([]db.CountExperimentVariantRunsSinceRow, error) {
	return s.counts[arg.WorkflowExperimentScopeID.Int32], nil
}

func (s *runtimeRoutingTestStore) ListActiveRoutingCompatibilitiesByInputSchemaID(context.Context, int32) ([]db.ListActiveRoutingCompatibilitiesByInputSchemaIDRow, error) {
	return nil, nil
}

func (s *runtimeRoutingTestStore) GetDefaultRouteForInputSchema(context.Context, int32) (db.WorkflowVersionInputSchemaCompatibility, error) {
	if s.defaultRoute.ID == 0 {
		return db.WorkflowVersionInputSchemaCompatibility{}, pgx.ErrNoRows
	}
	return s.defaultRoute, nil
}

func (s *runtimeRoutingTestStore) GetActiveWorkflowVersionInputSchemaCompatibilityByPair(_ context.Context, arg db.GetActiveWorkflowVersionInputSchemaCompatibilityByPairParams) (db.WorkflowVersionInputSchemaCompatibility, error) {
	route, ok := s.pairRoutes[[2]int32{arg.WorkflowVersionID, arg.WorkflowInputSchemaID}]
	if !ok {
		return db.WorkflowVersionInputSchemaCompatibility{}, pgx.ErrNoRows
	}
	return route, nil
}

func (s *runtimeRoutingTestStore) GetWorkflowInputMapperByID(context.Context, int32) (db.WorkflowInputMapper, error) {
	if s.mapper.ID == 0 {
		return db.WorkflowInputMapper{}, pgx.ErrNoRows
	}
	return s.mapper, nil
}

func TestBuildVersionInputDataAppliesMapperAndCompatibilityDefaults(t *testing.T) {
	store := &runtimeRoutingTestStore{
		mapper: db.WorkflowInputMapper{
			ID:         44,
			MapperType: "json",
			IsActive:   true,
			Rules: []byte(`{
				"copy_all": true,
				"mapping": {
					"recipient.email": "email",
					"metadata.plan": {"literal": "pro"}
				}
			}`),
		},
	}
	svc := NewService(store, client.Client(nil))

	got, err := svc.buildVersionInputData(context.Background(), store, runtimeRouteSelection{
		WorkflowVersionID:     12,
		WorkflowInputMapperID: pgtype.Int4{Int32: 44, Valid: true},
		CompatibilityType:     "adapter",
		DefaultValues:         []byte(`{"locale":"ru","priority":"normal"}`),
		SelectionReason:       "default_route",
		RoutingDecision:       []byte(`{}`),
	}, []byte(`{"email":"ada@example.com","locale":"en"}`))

	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(got, &decoded))
	require.Equal(t, "ada@example.com", decoded["email"])
	require.Equal(t, "en", decoded["locale"])
	require.Equal(t, "normal", decoded["priority"])

	recipient, ok := decoded["recipient"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "ada@example.com", recipient["email"])
	metadata, ok := decoded["metadata"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "pro", metadata["plan"])
}

func TestRouteSelectionSeedPrefersIdempotencyKey(t *testing.T) {
	got := routeSelectionSeed(map[string]any{"email": "ada@example.com"}, "stable-key", func() string {
		t.Fatal("random seed must not be used when idempotency key is present")
		return ""
	})

	require.Equal(t, "stable-key", got)
}

func TestRouteSelectionSeedUsesRandomSeedWhenIdempotencyKeyIsMissing(t *testing.T) {
	got := routeSelectionSeed(map[string]any{"email": "ada@example.com"}, "", func() string {
		return "request-seed"
	})

	require.Equal(t, "request-seed", got)
}

func TestSelectRuntimeRouteErrorsWhenExperimentFallbackPolicyIsError(t *testing.T) {
	store := &runtimeRoutingTestStore{
		scopes: []db.ListActiveExperimentScopesForRoutingRow{{
			WorkflowExperimentID:      7,
			WorkflowExperimentScopeID: 77,
			TrafficConditions:         []byte(`{}`),
			TrafficPercent:            100,
			FallbackPolicy:            "error",
		}},
		variants: map[int32][]db.WorkflowExperimentVariant{77: {}},
		defaultRoute: db.WorkflowVersionInputSchemaCompatibility{
			ID:                10,
			WorkflowVersionID: 20,
			CompatibilityType: "native",
			IsActive:          true,
			IsDefaultRoute:    true,
		},
	}
	svc := NewService(store, client.Client(nil))

	_, err := svc.selectRuntimeRoute(context.Background(), 1, 2, nil, []byte(`{"email":"ada@example.com"}`), "stable-key")

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "fallback policy is error"), err.Error())
}

func TestSelectRuntimeRouteUsesFallbackVersionWhenScopeTrafficIsExcluded(t *testing.T) {
	store := &runtimeRoutingTestStore{
		scopes: []db.ListActiveExperimentScopesForRoutingRow{{
			WorkflowExperimentID:      7,
			WorkflowExperimentScopeID: 77,
			TrafficConditions:         []byte(`{}`),
			TrafficPercent:            0,
			FallbackPolicy:            "fallback_version",
			FallbackWorkflowVersionID: pgtype.Int4{Int32: 22, Valid: true},
		}},
		variants: map[int32][]db.WorkflowExperimentVariant{77: {
			{ID: 88, WorkflowExperimentScopeID: 77, WorkflowVersionID: 33, TrafficWeight: 100, IsActive: true},
		}},
		pairRoutes: map[[2]int32]db.WorkflowVersionInputSchemaCompatibility{
			{22, 2}: {
				ID:                    99,
				WorkflowVersionID:     22,
				WorkflowInputSchemaID: 2,
				CompatibilityType:     "native",
				IsActive:              true,
			},
		},
	}
	svc := NewService(store, client.Client(nil))

	got, err := svc.selectRuntimeRoute(context.Background(), 1, 2, nil, []byte(`{"email":"ada@example.com"}`), "stable-key")

	require.NoError(t, err)
	require.Equal(t, int32(22), got.WorkflowVersionID)
	require.Equal(t, int32(99), got.InputSchemaCompatibilityID.Int32)
	require.Equal(t, int32(7), got.WorkflowExperimentID.Int32)
	require.Equal(t, int32(77), got.WorkflowExperimentScopeID.Int32)
	require.False(t, got.WorkflowExperimentVariantID.Valid)
	require.Equal(t, "fallback", got.SelectionReason)
}

func TestSelectRuntimeRouteUsesExplicitExperiment(t *testing.T) {
	expID := int32(7)
	store := &runtimeRoutingTestStore{
		scopes: []db.ListActiveExperimentScopesForRoutingRow{{
			WorkflowExperimentID:      7,
			WorkflowExperimentScopeID: 77,
			TrafficConditions:         []byte(`{}`),
			TrafficPercent:            100,
			FallbackPolicy:            "default_route",
		}},
		variants: map[int32][]db.WorkflowExperimentVariant{77: {
			{ID: 88, WorkflowExperimentScopeID: 77, WorkflowVersionID: 33, TrafficWeight: 100, IsActive: true},
		}},
		pairRoutes: map[[2]int32]db.WorkflowVersionInputSchemaCompatibility{
			{33, 2}: {ID: 99, WorkflowVersionID: 33, WorkflowInputSchemaID: 2, CompatibilityType: "native", IsActive: true},
		},
	}
	svc := NewService(store, client.Client(nil))

	got, err := svc.selectRuntimeRoute(context.Background(), 1, 2, &expID, []byte(`{"email":"ada@example.com"}`), "stable-key")

	require.NoError(t, err)
	require.True(t, store.routingParams.ExperimentID.Valid)
	require.Equal(t, int32(7), store.routingParams.ExperimentID.Int32)
	require.Equal(t, int32(7), got.WorkflowExperimentID.Int32)
	require.Equal(t, int32(88), got.WorkflowExperimentVariantID.Int32)
	require.Equal(t, int32(33), got.WorkflowVersionID)
}

func TestSelectRuntimeRouteChoosesMostUnderrepresentedVariantSinceExperimentTrafficChange(t *testing.T) {
	expID := int32(7)
	store := &runtimeRoutingTestStore{
		scopes: []db.ListActiveExperimentScopesForRoutingRow{{
			WorkflowExperimentID:      7,
			WorkflowExperimentScopeID: 77,
			TrafficConditions:         []byte(`{}`),
			TrafficPercent:            100,
			FallbackPolicy:            "default_route",
			TrafficChangedAt:          pgtype.Timestamp{Valid: true},
		}},
		variants: map[int32][]db.WorkflowExperimentVariant{77: {
			{ID: 88, WorkflowExperimentScopeID: 77, WorkflowVersionID: 33, TrafficWeight: 50, IsControlGroup: true, IsActive: true},
			{ID: 89, WorkflowExperimentScopeID: 77, WorkflowVersionID: 34, TrafficWeight: 50, IsActive: true},
		}},
		counts: map[int32][]db.CountExperimentVariantRunsSinceRow{77: {
			{WorkflowExperimentVariantID: 88, RunCount: 10},
			{WorkflowExperimentVariantID: 89, RunCount: 0},
		}},
		pairRoutes: map[[2]int32]db.WorkflowVersionInputSchemaCompatibility{
			{33, 2}: {ID: 99, WorkflowVersionID: 33, WorkflowInputSchemaID: 2, CompatibilityType: "native", IsActive: true},
			{34, 2}: {ID: 100, WorkflowVersionID: 34, WorkflowInputSchemaID: 2, CompatibilityType: "native", IsActive: true},
		},
	}
	svc := NewService(store, client.Client(nil))

	got, err := svc.selectRuntimeRoute(context.Background(), 1, 2, &expID, []byte(`{"email":"ada@example.com"}`), "request-1")

	require.NoError(t, err)
	require.Equal(t, int32(89), got.WorkflowExperimentVariantID.Int32)
	require.Equal(t, int32(34), got.WorkflowVersionID)
}

func TestSelectRuntimeRouteErrorsWhenExplicitExperimentHasNoActiveScope(t *testing.T) {
	expID := int32(7)
	store := &runtimeRoutingTestStore{scopes: []db.ListActiveExperimentScopesForRoutingRow{}}
	svc := NewService(store, client.Client(nil))

	_, err := svc.selectRuntimeRoute(context.Background(), 1, 2, &expID, []byte(`{"email":"ada@example.com"}`), "stable-key")

	require.Error(t, err)
	require.Contains(t, err.Error(), "EXPERIMENT_NOT_ROUTABLE")
}
