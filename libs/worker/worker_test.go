package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	raw := `{"success":true,"data":{"id":42,"work_type_id":1,"worker_settings_schema_id":3,"name":"smtp-worker","metadata":"e30=","registered_at":"2026-01-01T00:00:00Z","last_heartbeat_at":"2026-01-01T00:00:00Z"}}`

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
}

// TestSendRegistration verifies end-to-end: sendRegistration sends the correct
// JSON schema, parses the response, and populates w.workerID.
func TestSendRegistration(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":7,"work_type_id":1,"name":"test-worker"}}`))
	}))
	defer srv.Close()

	wk := &Worker{
		cfg: Config{
			ManagerURL:     srv.URL,
			BootstrapToken: "test-token",
			WorkerName:     "test-worker",
			Manifest: Manifest{
				Kind:        "email",
				NameKind:    "Email",
				Type:        "smtp",
				NameType:    "SMTP",
				InputSchema: json.RawMessage(`{}`),
			},
		},
		logger: slog.Default(),
	}

	if err := wk.sendRegistration(context.Background()); err != nil {
		t.Fatalf("sendRegistration returned error: %v", err)
	}

	if wk.workerID != int32(7) {
		t.Errorf("expected workerID == 7, got %d", wk.workerID)
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
	schema := ObjectSchema(
		Field("to", StringField(Required(), Description("Recipient email"))),
		Field("subject", StringField(Required())),
		Field("count", IntegerField()),
	)

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
}
