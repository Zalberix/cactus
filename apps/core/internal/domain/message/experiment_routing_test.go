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

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type runtimeRoutingTestStore struct {
	detailStore

	scopes       []db.ListActiveExperimentScopesForRoutingRow
	variants     map[int32][]db.WorkflowExperimentVariant
	defaultRoute db.WorkflowVersionInputSchemaCompatibility
	pairRoutes   map[[2]int32]db.WorkflowVersionInputSchemaCompatibility
	mapper       db.WorkflowInputMapper
}

func (s *runtimeRoutingTestStore) ListActiveExperimentScopesForRouting(context.Context, db.ListActiveExperimentScopesForRoutingParams) ([]db.ListActiveExperimentScopesForRoutingRow, error) {
	return s.scopes, nil
}

func (s *runtimeRoutingTestStore) ListActiveWorkflowExperimentVariantsByScopeID(_ context.Context, workflowExperimentScopeID int32) ([]db.WorkflowExperimentVariant, error) {
	return s.variants[workflowExperimentScopeID], nil
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

	_, err := svc.selectRuntimeRoute(context.Background(), 1, 2, []byte(`{"email":"ada@example.com"}`), "stable-key")

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

	got, err := svc.selectRuntimeRoute(context.Background(), 1, 2, []byte(`{"email":"ada@example.com"}`), "stable-key")

	require.NoError(t, err)
	require.Equal(t, int32(22), got.WorkflowVersionID)
	require.Equal(t, int32(99), got.InputSchemaCompatibilityID.Int32)
	require.Equal(t, int32(7), got.WorkflowExperimentID.Int32)
	require.Equal(t, int32(77), got.WorkflowExperimentScopeID.Int32)
	require.False(t, got.WorkflowExperimentVariantID.Valid)
	require.Equal(t, "experiment_scope_traffic_excluded_fallback_version", got.SelectionReason)
}
