package wshub

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/zalberix/cactus/apps/manager/internal/domain/auth"
	"github.com/zalberix/cactus/apps/manager/internal/domain/message"
	"github.com/zalberix/cactus/apps/manager/internal/http/response"
	"github.com/zalberix/cactus/apps/manager/internal/store"
	temporaltypes "github.com/zalberix/cactus/apps/manager/internal/temporal"
	"github.com/zalberix/cactus/libs/bus"
)

const (
	authTimeout              = 5 * time.Second  // per D-13
	pingInterval             = 30 * time.Second // per D-19
	defaultStatePollInterval = 2 * time.Second
	defaultEventFetchMaxWait = 500 * time.Millisecond
)

// Hub manages WebSocket connections for real-time workflow status.
type Hub struct {
	msgService messageDetailService
	authSvc    *auth.Service
	store      *store.Store // for system token validation
	events     workflowEventSource
	logger     *slog.Logger

	statePollInterval time.Duration
	eventFetchMaxWait time.Duration
}

// New creates a new Hub.
func New(b *bus.Bus, msgSvc *message.Service, authSvc *auth.Service, s *store.Store) *Hub {
	return &Hub{
		msgService: msgSvc,
		authSvc:    authSvc,
		store:      s,
		events:     jetstreamWorkflowEventSource{bus: b, logger: slog.Default()},
		logger:     slog.Default(),
	}
}

type messageDetailService interface {
	GetMessageDetail(ctx context.Context, messageID int32) (*message.DetailResponse, error)
}

type workflowEventSource interface {
	SubscribeWorkflow(ctx context.Context, messageID int32) (workflowEventSubscription, error)
}

type workflowEventSubscription interface {
	Fetch(ctx context.Context, maxWait time.Duration) (temporaltypes.WorkflowEvent, bool, error)
}

type jetstreamWorkflowEventSource struct {
	bus    *bus.Bus
	logger *slog.Logger
}

func (s jetstreamWorkflowEventSource) SubscribeWorkflow(ctx context.Context, messageID int32) (workflowEventSubscription, error) {
	subject := workflowEventSubject(messageID)
	cons, err := s.bus.JS().CreateConsumer(ctx, "EVENTS", jetstream.ConsumerConfig{
		FilterSubject:     subject,
		AckPolicy:         jetstream.AckExplicitPolicy,
		DeliverPolicy:     jetstream.DeliverNewPolicy, // only new events, snapshot covers history
		InactiveThreshold: 2 * time.Minute,
	})
	if err != nil {
		return nil, err
	}
	return jetstreamWorkflowEventSubscription{consumer: cons, logger: s.logger}, nil
}

type jetstreamWorkflowEventSubscription struct {
	consumer jetstream.Consumer
	logger   *slog.Logger
}

func (s jetstreamWorkflowEventSubscription) Fetch(ctx context.Context, maxWait time.Duration) (temporaltypes.WorkflowEvent, bool, error) {
	if err := ctx.Err(); err != nil {
		return temporaltypes.WorkflowEvent{}, false, err
	}

	msgs, err := s.consumer.Fetch(1, jetstream.FetchMaxWait(maxWait))
	if err != nil {
		return temporaltypes.WorkflowEvent{}, false, err
	}

	for natsMsg := range msgs.Messages() {
		_ = natsMsg.Ack()

		var event temporaltypes.WorkflowEvent
		if err := json.Unmarshal(natsMsg.Data(), &event); err != nil {
			if s.logger != nil {
				s.logger.Warn("unmarshal workflow event failed", slog.String("error", err.Error()))
			}
			return temporaltypes.WorkflowEvent{}, false, nil
		}
		return event, true, nil
	}

	if err := msgs.Error(); err != nil {
		return temporaltypes.WorkflowEvent{}, false, err
	}
	return temporaltypes.WorkflowEvent{}, false, nil
}

func workflowEventSubject(messageID int32) string {
	return fmt.Sprintf("event.workflow.%d", messageID)
}

func (h *Hub) effectiveStatePollInterval() time.Duration {
	if h.statePollInterval > 0 {
		return h.statePollInterval
	}
	return defaultStatePollInterval
}

func (h *Hub) effectiveEventFetchMaxWait() time.Duration {
	if h.eventFetchMaxWait > 0 {
		return h.eventFetchMaxWait
	}
	return defaultEventFetchMaxWait
}

// HandleWS is the Gin handler for WebSocket connections at /ws/workflow/:messageID (per D-12).
//
// Protocol:
//  1. Extract messageID from URL param; 400 if invalid (pre-upgrade).
//  2. Upgrade to WebSocket.
//  3. Auth handshake: read first message within 5s, validate JWT or system token (per D-13).
//  4. Send auth_ok.
//  5. Subscribe to NATS event.workflow.{messageID}.
//  6. Build and send snapshot from DB (per D-14).
//  7. If workflow already done/failed, send terminal event and close (per D-18).
//  8. Forward buffered and new delta events.
//  9. Ping/pong heartbeat every 30s (per D-19).
func (h *Hub) HandleWS(c *gin.Context) { //nolint:gocognit // WS lifecycle keeps auth, snapshot, subscription, and ping handling together.
	// 1. Extract and validate messageID (pre-upgrade).
	messageIDStr := c.Param("messageID")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 32)
	if err != nil || messageID <= 0 {
		response.BadRequest(c, "INVALID_MESSAGE_ID", "messageID must be a positive integer")
		return
	}
	msgID := int32(messageID)

	// 2. Upgrade to WebSocket.
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{})
	if err != nil {
		h.logger.Error("websocket accept failed", slog.String("error", err.Error()))
		return
	}
	defer conn.CloseNow()

	ctx := c.Request.Context()

	// 3. Auth handshake (per D-13): read first message within 5s.
	authCtx, authCancel := context.WithTimeout(ctx, authTimeout)
	defer authCancel()

	var authMsg AuthMessage
	if err := wsjson.Read(authCtx, conn, &authMsg); err != nil {
		conn.Close(websocket.StatusPolicyViolation, "auth timeout or read error")
		return
	}

	if authMsg.Type != "auth" || authMsg.Token == "" {
		conn.Close(websocket.StatusPolicyViolation, "invalid auth message")
		return
	}

	if err := h.authenticateWS(ctx, authMsg.Token); err != nil {
		h.logger.Debug("ws auth failed", slog.String("error", err.Error()))
		conn.Close(websocket.StatusPolicyViolation, "invalid token")
		return
	}

	// 4. Send auth_ok.
	if err := wsjson.Write(ctx, conn, AuthOK{Type: "auth_ok"}); err != nil {
		return
	}

	// 5. Subscribe before snapshot so events that happen during the DB read are buffered.
	eventSub, err := h.events.SubscribeWorkflow(ctx, msgID)
	if err != nil {
		h.logger.Error("create NATS consumer failed", slog.String("error", err.Error()), slog.String("subject", workflowEventSubject(msgID)))
		conn.Close(websocket.StatusInternalError, "event subscription failed")
		return
	}

	// 6. Snapshot (per D-14): query current detail from DB.
	detail, err := h.msgService.GetMessageDetail(ctx, msgID)
	if err != nil {
		h.logger.Error("get message detail failed", slog.String("error", err.Error()), slog.Int("messageID", int(msgID)))
		conn.Close(websocket.StatusInternalError, "message not found")
		return
	}

	snapshot := h.buildSnapshot(detail)
	if err := wsjson.Write(ctx, conn, snapshot); err != nil {
		return
	}

	// 7. Terminal check (per D-18): if already done/failed, send terminal event and close.
	if detail.WorkflowRun != nil &&
		(detail.WorkflowRun.Status == temporaltypes.RunStatusCompleted ||
			detail.WorkflowRun.Status == temporaltypes.RunStatusFailed) {
		h.sendTerminalEvent(ctx, conn, detail.WorkflowRun.Status, detail)
		conn.Close(websocket.StatusNormalClosure, "workflow finished")
		return
	}

	// 8. Event loop: readPump + writePump.
	done := make(chan struct{})

	// readPump: detect client disconnect.
	go func() {
		defer close(done)
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				return
			}
			// Ignore any client messages after auth (WS is read-only in v1).
		}
	}()

	// writePump: ping + NATS event forwarding.
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()
	stateTicker := time.NewTicker(h.effectiveStatePollInterval())
	defer stateTicker.Stop()

	for {
		select {
		case <-done:
			// Client disconnected.
			return

		case <-ctx.Done():
			// Server shutting down.
			conn.Close(websocket.StatusGoingAway, "server shutdown")
			return

		case <-pingTicker.C:
			// Per D-19: server-initiated ping.
			pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
			if err := conn.Ping(pingCtx); err != nil {
				pingCancel()
				return
			}
			pingCancel()

		case <-stateTicker.C:
			terminal, err := h.sendSnapshotFromDB(ctx, conn, msgID)
			if err != nil {
				return
			}
			if terminal {
				conn.Close(websocket.StatusNormalClosure, "workflow finished")
				return
			}

		default:
			// Poll for NATS messages with short timeout.
			event, ok, err := eventSub.Fetch(ctx, h.effectiveEventFetchMaxWait())
			if err != nil {
				// Timeout or context cancel — continue loop.
				continue
			}

			if !ok {
				continue
			}

			if isTerminalWorkflowEvent(event) {
				terminal, err := h.sendSnapshotFromDB(ctx, conn, msgID)
				if err != nil {
					return
				}
				if !terminal {
					if err := wsjson.Write(ctx, conn, event); err != nil {
						return
					}
				}
				conn.Close(websocket.StatusNormalClosure, "workflow finished")
				return
			}

			// Forward event to WS client as-is (same JSON structure).
			if err := wsjson.Write(ctx, conn, event); err != nil {
				return
			}
		}
	}
}

func isTerminalWorkflowEvent(event temporaltypes.WorkflowEvent) bool {
	return event.Type == "workflow_done" || event.Type == "workflow_failed"
}

func (h *Hub) sendSnapshotFromDB(ctx context.Context, conn *websocket.Conn, messageID int32) (bool, error) {
	detail, err := h.msgService.GetMessageDetail(ctx, messageID)
	if err != nil {
		h.logger.Warn("refresh message detail failed", slog.String("error", err.Error()), slog.Int("messageID", int(messageID)))
		return false, nil
	}
	if err := wsjson.Write(ctx, conn, h.buildSnapshot(detail)); err != nil {
		return false, err
	}
	if isTerminalDetail(detail) {
		h.sendTerminalEvent(ctx, conn, detail.WorkflowRun.Status, detail)
		return true, nil
	}
	return false, nil
}

func isTerminalDetail(detail *message.DetailResponse) bool {
	return detail != nil &&
		detail.WorkflowRun != nil &&
		(detail.WorkflowRun.Status == temporaltypes.RunStatusCompleted ||
			detail.WorkflowRun.Status == temporaltypes.RunStatusFailed)
}

// buildSnapshot constructs a SnapshotEvent from the DB detail response.
func (h *Hub) buildSnapshot(detail *message.DetailResponse) SnapshotEvent {
	return SnapshotEvent{
		Type:   "snapshot",
		Detail: detail,
	}
}

// sendTerminalEvent sends the appropriate terminal event based on workflow status.
func (h *Hub) sendTerminalEvent(ctx context.Context, conn *websocket.Conn, status string, detail *message.DetailResponse) {
	now := time.Now().UTC().Format(time.RFC3339)
	_ = wsjson.Write(ctx, conn, terminalEventForStatus(status, detail, now))
}

func terminalEventForStatus(status string, detail *message.DetailResponse, timestamp string) any {
	switch status {
	case temporaltypes.RunStatusCompleted:
		return WorkflowDoneEvent{
			Type:      "workflow_done",
			Timestamp: timestamp,
		}
	case temporaltypes.RunStatusFailed:
		errMsg := ""
		if detail != nil && detail.WorkflowRun != nil && detail.WorkflowRun.ErrorMessage != nil {
			errMsg = *detail.WorkflowRun.ErrorMessage
		}
		return WorkflowFailedEvent{
			Type:      "workflow_failed",
			Error:     errMsg,
			Timestamp: timestamp,
		}
	}
	return map[string]string{"type": "workflow_unknown", "timestamp": timestamp}
}

// authenticateWS validates the auth token from the WS handshake.
// Per D-13: token is either a JWT string or a system token in
// "public_token:private_token" colon-separated format.
// Returns nil on success, error on failure.
func (h *Hub) authenticateWS(ctx context.Context, token string) error {
	// 1. Try JWT first.
	_, err := h.authSvc.ParseToken(token)
	if err == nil {
		return nil // JWT valid
	}

	// 2. Try system token: expect "public_token:private_token" format.
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid token: not a valid JWT and not in public:private format")
	}

	publicToken, privateToken := parts[0], parts[1]

	dbToken, err := h.store.GetSystemTokenByPublicToken(ctx, publicToken)
	if err != nil {
		return fmt.Errorf("system token lookup failed: %w", err)
	}

	// Hash incoming private token and compare (constant-time).
	incomingHash := sha256hex(privateToken)
	if subtle.ConstantTimeCompare([]byte(dbToken.PrivateToken), []byte(incomingHash)) != 1 {
		return fmt.Errorf("invalid system token: hash mismatch")
	}

	return nil // system token valid
}

// sha256hex computes SHA256 hash and returns hex string.
// Same algorithm as middleware.sha256hex (duplicated because unexported).
func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
