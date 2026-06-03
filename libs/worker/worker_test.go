package worker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// TestRegisterRequestJSON verifies that registerRequest marshals to JSON
// matching the API contract (RegisterWorkerRequest): bootstrap_token, name, manifest.
func TestRegisterRequestJSON(t *testing.T) {
	req := registerRequest{
		BootstrapToken: "tok123",
		Name:           "smtp-dev",
		Manifest:       json.RawMessage(`{"kind":"email","name_kind":"Email","type":"smtp","name_type":"SMTP","input_schema":{}}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal registerRequest: %v", err)
	}

	// Decode top-level keys to verify schema
	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(data, &topLevel); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	// Must contain correct top-level keys
	for _, key := range []string{"bootstrap_token", "name", "manifest"} {
		if _, ok := topLevel[key]; !ok {
			t.Errorf("JSON must have top-level key %q, got keys: %v", key, topLevelKeys(topLevel))
		}
	}

	// Verify values
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"bootstrap_token":"tok123"`) {
		t.Errorf("expected bootstrap_token value 'tok123', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"name":"smtp-dev"`) {
		t.Errorf("expected name value 'smtp-dev', got: %s", jsonStr)
	}

	// Must NOT contain old field names as top-level keys
	oldKeys := []string{"token", "worker_uuid", "kind", "name_kind", "type", "name_type", "config_schema"}
	for _, key := range oldKeys {
		if _, ok := topLevel[key]; ok {
			t.Errorf("JSON must NOT have top-level key %q", key)
		}
	}
}

func topLevelKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestRegisterResponseUnmarshal verifies that registerResponse correctly parses
// the Manager API envelope, extracting Data.ID (db.Worker.ID).
func TestRegisterResponseUnmarshal(t *testing.T) {
	raw := `{"success":true,"data":{"id":42,"organization_id":9,"work_type_id":12,"revision_id":99,"worker_settings_schema_id":3,"name":"smtp-worker","nats":{"url":"tls://localhost:4222","user_jwt":"jwt","user_seed":"seed","credentials":"creds"},"metadata":"e30=","registered_at":"2026-01-01T00:00:00Z","last_heartbeat_at":"2026-01-01T00:00:00Z"}}`

	var resp registerResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal registerResponse: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success == true")
	}
	if resp.Data.ID != int32(42) {
		t.Errorf("expected Data.ID == 42, got %d", resp.Data.ID)
	}
	if resp.Data.OrganizationID != int32(9) {
		t.Errorf("expected Data.OrganizationID == 9, got %d", resp.Data.OrganizationID)
	}
	if resp.Data.WorkTypeID != int32(12) {
		t.Errorf("expected Data.WorkTypeID == 12, got %d", resp.Data.WorkTypeID)
	}
	if resp.Data.WorkerSettingsSchemaID != int32(3) {
		t.Errorf("expected Data.WorkerSettingsSchemaID == 3, got %d", resp.Data.WorkerSettingsSchemaID)
	}
	if resp.Data.RevisionID != int32(99) {
		t.Errorf("expected Data.RevisionID == 99, got %d", resp.Data.RevisionID)
	}
	if resp.Data.NATS.UserJWT != "jwt" || resp.Data.NATS.UserSeed != "seed" {
		t.Errorf("expected NATS credentials, got %#v", resp.Data.NATS)
	}
}

// TestSendRegistration verifies end-to-end: sendRegistration sends the correct
// JSON schema, parses the response, and populates w.workerID.
func TestSendRegistration(t *testing.T) { //nolint:gocognit // End-to-end request/response assertions stay together for readability.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/register/worker" {
			t.Errorf("expected path /api/v1/register/worker, got %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		// Decode body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}

		// Assert required keys
		if bt, ok := m["bootstrap_token"]; !ok {
			t.Error("missing key 'bootstrap_token'")
		} else if bt != "test-token" {
			t.Errorf("expected bootstrap_token 'test-token', got %v", bt)
		}

		if name, ok := m["name"]; !ok {
			t.Error("missing key 'name'")
		} else if name != "test-worker" {
			t.Errorf("expected name 'test-worker', got %v", name)
		}

		manifest, ok := m["manifest"]
		if !ok {
			t.Error("missing key 'manifest'")
			return
		}

		// manifest must be a JSON object (map)
		mf, isMap := manifest.(map[string]any)
		if !isMap {
			t.Errorf("expected manifest to be an object, got %T", manifest)
			return
		}
		for _, key := range []string{"kind", "name_kind", "type", "name_type", "input_schema"} {
			if _, exists := mf[key]; !exists {
				t.Errorf("manifest missing key %q", key)
			}
		}

		// Respond with envelope
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":7,"organization_id":9,"work_type_id":12,"worker_settings_schema_id":3,"revision_id":99,"name":"test-worker","nats":{"url":"tls://localhost:4222","user_jwt":"jwt","user_seed":"seed","credentials":"creds"}}}`))
	}))
	defer srv.Close()

	wk := &Worker{
		cfg: Config{
			ManagerURL:     srv.URL,
			BootstrapToken: "test-token",
			WorkerName:     "test-worker",
			Manifest: Manifest().
				Kind("email", "Email").
				Type("smtp", "SMTP").
				InputSchema(func(*SchemaBuilder) {}).
				Build(),
		},
		logger: slog.Default(),
	}

	if err := wk.sendRegistration(context.Background()); err != nil {
		t.Fatalf("sendRegistration returned error: %v", err)
	}

	if wk.workerID != int32(7) {
		t.Errorf("expected workerID == 7, got %d", wk.workerID)
	}
	if wk.cfg.WorkTypeID != int32(12) {
		t.Errorf("expected runtime WorkTypeID == 12, got %d", wk.cfg.WorkTypeID)
	}
	if wk.cfg.OrganizationID != int32(9) {
		t.Errorf("expected runtime OrganizationID == 9, got %d", wk.cfg.OrganizationID)
	}
	if wk.cfg.WorkerSettingsSchemaID != int32(3) {
		t.Errorf("expected runtime WorkerSettingsSchemaID == 3, got %d", wk.cfg.WorkerSettingsSchemaID)
	}
	if wk.cfg.RevisionID != int32(99) {
		t.Errorf("expected runtime RevisionID == 99, got %d", wk.cfg.RevisionID)
	}
	if wk.natsCreds.UserJWT != "jwt" || wk.natsCreds.UserSeed != "seed" {
		t.Errorf("expected runtime NATS credentials, got %#v", wk.natsCreds)
	}
}

func TestRequireRoutingConfiguredRejectsMissingRegistrationIDs(t *testing.T) {
	wk := &Worker{}

	err := wk.requireRoutingConfigured()
	if err == nil {
		t.Fatal("expected missing routing IDs error")
	}
	if !strings.Contains(err.Error(), "registration response missing organization_id, work_type_id, worker_settings_schema_id or revision_id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskFilterSubjectIncludesSettingsSchema(t *testing.T) {
	got := taskFilterSubject(12, 3, 7)
	want := "task.org.12.work_type.3.schema.7.>"
	if got != want {
		t.Fatalf("taskFilterSubject() = %q, want %q", got, want)
	}
}

func TestTaskConsumerNameIsSharedBySchema(t *testing.T) {
	got := taskConsumerName(12, 3, 7)
	want := "task-org-12-work-type-3-schema-7"
	if got != want {
		t.Fatalf("consumer name = %q, want %q", got, want)
	}
}

func TestTaskConsumerConfigUsesSharedSchemaConsumer(t *testing.T) {
	cfg := taskConsumerConfig(12, 3, 7)

	if cfg.Name != "task-org-12-work-type-3-schema-7" {
		t.Fatalf("name = %q", cfg.Name)
	}
	if cfg.Durable != cfg.Name {
		t.Fatalf("durable = %q, want %q", cfg.Durable, cfg.Name)
	}
	if cfg.FilterSubject != "task.org.12.work_type.3.schema.7.>" {
		t.Fatalf("filter = %q", cfg.FilterSubject)
	}
	if cfg.AckPolicy != jetstream.AckExplicitPolicy {
		t.Fatalf("ack policy = %v", cfg.AckPolicy)
	}
	if cfg.MaxDeliver != 5 {
		t.Fatalf("max deliver = %d", cfg.MaxDeliver)
	}
}

func TestTaskConsumerConfigDoesNotExpireDurableWorker(t *testing.T) {
	cfg := taskConsumerConfig(12, 3, 7)

	if cfg.InactiveThreshold != 0 {
		t.Fatalf("durable worker consumer must not expire while worker is alive, got %s", cfg.InactiveThreshold)
	}
}

func TestNewDefaultsTaskTimeout(t *testing.T) {
	w := New(Config{}, nil, nil)

	if w.cfg.TaskTimeout != defaultTaskTimeout {
		t.Fatalf("task timeout = %s, want %s", w.cfg.TaskTimeout, defaultTaskTimeout)
	}
}

func TestNewKeepsConfiguredTaskTimeout(t *testing.T) {
	w := New(Config{TaskTimeout: 3 * time.Second}, nil, nil)

	if w.cfg.TaskTimeout != 3*time.Second {
		t.Fatalf("task timeout = %s, want 3s", w.cfg.TaskTimeout)
	}
}

func TestNATSConnectOptionsRetryForeverEverySecond(t *testing.T) {
	w := New(Config{}, nil, nil)
	w.natsCreds = NATSCredentials{
		URL:      "tls://localhost:4222",
		UserJWT:  "jwt",
		UserSeed: "seed",
	}

	opts := nats.GetDefaultOptions()
	for _, opt := range w.natsConnectOptions() {
		if err := opt(&opts); err != nil {
			t.Fatalf("apply option: %v", err)
		}
	}

	if opts.MaxReconnect != -1 {
		t.Fatalf("MaxReconnect = %d, want -1", opts.MaxReconnect)
	}
	if opts.ReconnectWait != time.Second {
		t.Fatalf("ReconnectWait = %s, want 1s", opts.ReconnectWait)
	}
}

func TestConsumeReturnsWhenNATSConnectionIsClosed(t *testing.T) {
	w := New(Config{}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	consumer := &fetchErrorConsumer{err: nats.ErrConnectionClosed}
	errCh := make(chan error, 1)

	go func() {
		errCh <- w.consume(context.Background(), consumer)
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, nats.ErrConnectionClosed) {
			t.Fatalf("consume returned %v, want nats.ErrConnectionClosed", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("consume did not return after closed NATS connection")
	}
}

func TestLoadOrRegisterUpdatesWorkerIDFileWhenManagerReturnsDifferentID(t *testing.T) {
	idPath := filepath.Join(t.TempDir(), "worker-id")
	if err := os.WriteFile(idPath, []byte("7"), 0o600); err != nil {
		t.Fatalf("write existing worker id: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":99,"work_type_id":1,"name":"test-worker"}}`))
	}))
	defer srv.Close()

	wk := &Worker{
		cfg: Config{
			ManagerURL:     srv.URL,
			BootstrapToken: "test-token",
			WorkerName:     "test-worker",
			WorkerIDPath:   idPath,
			Manifest: Manifest().
				Kind("email", "Email").
				Type("smtp", "SMTP").
				InputSchema(func(*SchemaBuilder) {}).
				Build(),
		},
		logger: slog.Default(),
	}

	if err := wk.loadOrRegister(context.Background()); err != nil {
		t.Fatalf("loadOrRegister returned error: %v", err)
	}

	data, err := os.ReadFile(idPath)
	if err != nil {
		t.Fatalf("read worker id file: %v", err)
	}
	if strings.TrimSpace(string(data)) != "99" {
		t.Fatalf("expected worker id file to be updated to 99, got %q", string(data))
	}
}

func TestRefreshLoggerAfterWorkerIDUsesCallback(t *testing.T) {
	callbackCalled := false
	expectedLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := New(Config{
		OnWorkerID: func(workerID int32) *slog.Logger {
			callbackCalled = true
			if workerID != 42 {
				t.Fatalf("OnWorkerID workerID = %d, want 42", workerID)
			}
			return expectedLogger
		},
	}, nil, slog.Default())
	w.workerID = 42

	w.refreshLoggerAfterWorkerID()

	if !callbackCalled {
		t.Fatal("expected OnWorkerID callback to be called")
	}
	if w.logger != expectedLogger {
		t.Fatal("expected worker logger to be replaced with callback logger")
	}
}

func TestSchemaBuilderStoresRequiredOnProperties(t *testing.T) {
	schema := SettingsSchema(func(ss *SchemaBuilder) {
		ss.String("to").Required().Description("Recipient email")
		ss.String("subject").Required()
		ss.Integer("count")
		ss.String("delivery").Enum("smtp", "telegram")
	})

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if _, exists := got["required"]; exists {
		t.Fatalf("schema must not contain top-level required: %s", data)
	}

	props := got["properties"].(map[string]any)
	to := props["to"].(map[string]any)
	if to["required"] != true {
		t.Fatalf("expected to.required=true, got %#v", to["required"])
	}
	if to["description"] != "Recipient email" {
		t.Fatalf("expected description to be preserved, got %#v", to["description"])
	}

	count := props["count"].(map[string]any)
	if _, exists := count["required"]; exists {
		t.Fatalf("optional field must not contain required: %#v", count)
	}

	delivery := props["delivery"].(map[string]any)
	enumValues := delivery["enum"].([]any)
	if len(enumValues) != 2 || enumValues[0] != "smtp" || enumValues[1] != "telegram" {
		t.Fatalf("expected enum to be preserved, got %#v", enumValues)
	}
}

func TestSchemaBuilderObjectField(t *testing.T) {
	schema := InputSchema(func(sb *SchemaBuilder) {
		sb.String("template").Required()
		sb.Object("fields").Required().Description("Template values as a JSON object")
	})

	var decoded map[string]any
	if err := json.Unmarshal(schema, &decoded); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	props := decoded["properties"].(map[string]any)
	fields := props["fields"].(map[string]any)
	if fields["type"] != "object" {
		t.Fatalf("expected fields type object, got %#v", fields)
	}
	if fields["required"] != true {
		t.Fatalf("expected fields to be required, got %#v", fields)
	}
	if fields["description"] != "Template values as a JSON object" {
		t.Fatalf("expected fields description, got %#v", fields)
	}
}

func TestManifestBuilderBuildsManifest(t *testing.T) {
	manifest := Manifest().
		Kind("telegram", "Telegram Bot").
		Type("social", "Social Delivery").
		SettingsSchema(func(ss *SchemaBuilder) {
			ss.String("server_url").Required()
		}).
		InputSchema(func(ss *SchemaBuilder) {
			ss.String("message").Required()
		}).
		OutputSchema(func(ss *SchemaBuilder) {
			ss.String("sent_at")
		}).
		Build()

	if manifest.Kind != "telegram" {
		t.Fatalf("expected kind telegram, got %q", manifest.Kind)
	}
	if manifest.NameKind != "Telegram Bot" {
		t.Fatalf("expected name kind Telegram Bot, got %q", manifest.NameKind)
	}
	if manifest.Type != "social" {
		t.Fatalf("expected type social, got %q", manifest.Type)
	}
	if manifest.NameType != "Social Delivery" {
		t.Fatalf("expected name type Social Delivery, got %q", manifest.NameType)
	}

	var settings map[string]any
	if err := json.Unmarshal(manifest.SettingsSchema, &settings); err != nil {
		t.Fatalf("unmarshal settings schema: %v", err)
	}

	properties := settings["properties"].(map[string]any)
	serverURL := properties["server_url"].(map[string]any)
	if serverURL["type"] != "string" {
		t.Fatalf("expected server_url type string, got %#v", serverURL["type"])
	}
	if serverURL["required"] != true {
		t.Fatalf("expected server_url required true, got %#v", serverURL["required"])
	}
}

func TestSelectVariantReturnsKnownVariant(t *testing.T) {
	handler := TaskHandler(nil)
	manifest := Manifest().Kind("smtp", "SMTP").Type("email", "Email").Build()
	variants := map[string]Variant{
		"basic": {
			Name:     "basic",
			Manifest: manifest,
			Handler:  handler,
		},
	}

	got, err := SelectVariant(variants, "basic")
	if err != nil {
		t.Fatalf("SelectVariant returned error: %v", err)
	}
	if got.Name != "basic" {
		t.Fatalf("expected basic, got %#v", got)
	}
	if got.Manifest.Kind != "smtp" {
		t.Fatalf("expected smtp manifest, got %#v", got.Manifest)
	}
}

func TestSelectVariantRejectsUnknownVariant(t *testing.T) {
	_, err := SelectVariant(map[string]Variant{"basic": {Name: "basic"}}, "auth")
	if err == nil {
		t.Fatal("expected unknown variant error")
	}
	if !strings.Contains(err.Error(), `unknown worker variant "auth"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

type fetchErrorConsumer struct {
	err error
}

func (c *fetchErrorConsumer) Fetch(int, ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) FetchBytes(int, ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) FetchNoWait(int) (jetstream.MessageBatch, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) Consume(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) Messages(...jetstream.PullMessagesOpt) (jetstream.MessagesContext, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) Next(...jetstream.FetchOpt) (jetstream.Msg, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) Info(context.Context) (*jetstream.ConsumerInfo, error) {
	return nil, c.err
}

func (c *fetchErrorConsumer) CachedInfo() *jetstream.ConsumerInfo {
	return nil
}
