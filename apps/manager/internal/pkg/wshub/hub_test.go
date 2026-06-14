package wshub

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/zalberix/cactus/apps/manager/config"
	"github.com/zalberix/cactus/apps/manager/internal/domain/auth"
	"github.com/zalberix/cactus/apps/manager/internal/domain/message"
	temporaltypes "github.com/zalberix/cactus/apps/manager/internal/temporal"
)

func TestBuildSnapshotUsesMessageDetail(t *testing.T) {
	detail := &message.DetailResponse{
		MessageID: 100,
		WorkflowRun: &message.WorkflowRunStatus{
			ID:     91,
			Status: temporaltypes.RunStatusFailed,
		},
	}

	snapshot := (&Hub{}).buildSnapshot(detail)
	if snapshot.Type != "snapshot" {
		t.Fatalf("type = %q", snapshot.Type)
	}
	if snapshot.Detail != detail {
		t.Fatalf("snapshot detail was not preserved")
	}
}

func TestTerminalFailedEventIncludesRunError(t *testing.T) {
	errMsg := "smtp failed"
	detail := &message.DetailResponse{
		WorkflowRun: &message.WorkflowRunStatus{
			Status:       temporaltypes.RunStatusFailed,
			ErrorMessage: &errMsg,
		},
	}

	event := terminalEventForStatus(temporaltypes.RunStatusFailed, detail, "2026-05-18T08:00:00Z")
	failed, ok := event.(WorkflowFailedEvent)
	if !ok {
		t.Fatalf("event type = %T", event)
	}
	if failed.Error != errMsg {
		t.Fatalf("error = %q", failed.Error)
	}
}

func TestHandleWSSubscribesBeforeSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := &callRecorder{}
	authSvc := auth.NewService(config.JWT{
		Secret:     "test-secret",
		AccessTTL:  time.Hour,
		RefreshTTL: time.Hour,
	}, nil)
	tokens, err := authSvc.IssueTokens(1)
	require.NoError(t, err)

	hub := &Hub{
		msgService: &recordingDetailService{recorder: recorder},
		authSvc:    authSvc,
		events:     &recordingEventSource{recorder: recorder},
		logger:     slog.Default(),
	}

	router := gin.New()
	router.GET("/ws/workflow/:messageID", hub.HandleWS)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/workflow/42"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	require.NoError(t, wsjson.Write(ctx, conn, AuthMessage{
		Type:  "auth",
		Token: tokens.AccessToken,
	}))

	var authOK AuthOK
	require.NoError(t, wsjson.Read(ctx, conn, &authOK))
	require.Equal(t, "auth_ok", authOK.Type)

	var snapshot SnapshotEvent
	require.NoError(t, wsjson.Read(ctx, conn, &snapshot))
	require.Equal(t, "snapshot", snapshot.Type)

	require.Equal(t, []string{"subscribe", "detail"}, recorder.calls())
}

func TestHandleWSSendsSnapshotFromDBWhenNoWorkflowEventsArrive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := auth.NewService(config.JWT{
		Secret:     "test-secret",
		AccessTTL:  time.Hour,
		RefreshTTL: time.Hour,
	}, nil)
	tokens, err := authSvc.IssueTokens(1)
	require.NoError(t, err)

	hub := &Hub{
		msgService:        &changingDetailService{},
		authSvc:           authSvc,
		events:            &recordingEventSource{recorder: &callRecorder{}},
		logger:            slog.Default(),
		statePollInterval: 20 * time.Millisecond,
		eventFetchMaxWait: 5 * time.Millisecond,
	}

	router := gin.New()
	router.GET("/ws/workflow/:messageID", hub.HandleWS)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/workflow/42"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	require.NoError(t, wsjson.Write(ctx, conn, AuthMessage{
		Type:  "auth",
		Token: tokens.AccessToken,
	}))

	var authOK AuthOK
	require.NoError(t, wsjson.Read(ctx, conn, &authOK))

	var initial SnapshotEvent
	require.NoError(t, wsjson.Read(ctx, conn, &initial))
	require.Equal(t, temporaltypes.RunStatusRunning, initial.Detail.WorkflowRun.Status)

	var refreshed SnapshotEvent
	require.NoError(t, wsjson.Read(ctx, conn, &refreshed))
	require.Equal(t, "snapshot", refreshed.Type)
	require.Equal(t, temporaltypes.RunStatusCompleted, refreshed.Detail.WorkflowRun.Status)
}

type callRecorder struct {
	mu      sync.Mutex
	entries []string
}

func (r *callRecorder) add(call string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, call)
}

func (r *callRecorder) calls() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.entries...)
}

type recordingDetailService struct {
	recorder *callRecorder
}

func (s *recordingDetailService) GetMessageDetail(_ context.Context, messageID int32) (*message.DetailResponse, error) {
	s.recorder.add("detail")
	return &message.DetailResponse{
		MessageID: messageID,
		WorkflowRun: &message.WorkflowRunStatus{
			ID:     7,
			Status: temporaltypes.RunStatusRunning,
		},
		Graph: message.GraphDTO{
			Steps:        []message.GraphStepDTO{},
			Dependencies: []message.GraphDependencyDTO{},
		},
		RunSteps: []message.StepRunDetailDTO{},
	}, nil
}

type recordingEventSource struct {
	recorder *callRecorder
}

func (s *recordingEventSource) SubscribeWorkflow(_ context.Context, messageID int32) (workflowEventSubscription, error) {
	s.recorder.add("subscribe")
	return idleWorkflowEventSubscription{}, nil
}

type idleWorkflowEventSubscription struct{}

func (idleWorkflowEventSubscription) Fetch(ctx context.Context, maxWait time.Duration) (temporaltypes.WorkflowEvent, bool, error) {
	timer := time.NewTimer(maxWait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return temporaltypes.WorkflowEvent{}, false, ctx.Err()
	case <-timer.C:
		return temporaltypes.WorkflowEvent{}, false, nil
	}
}

type changingDetailService struct {
	mu    sync.Mutex
	calls int
}

func (s *changingDetailService) GetMessageDetail(_ context.Context, messageID int32) (*message.DetailResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++
	status := temporaltypes.RunStatusRunning
	if s.calls > 1 {
		status = temporaltypes.RunStatusCompleted
	}
	return &message.DetailResponse{
		MessageID: messageID,
		WorkflowRun: &message.WorkflowRunStatus{
			ID:     7,
			Status: status,
		},
		Graph: message.GraphDTO{
			Steps:        []message.GraphStepDTO{},
			Dependencies: []message.GraphDependencyDTO{},
		},
		RunSteps: []message.StepRunDetailDTO{},
	}, nil
}
