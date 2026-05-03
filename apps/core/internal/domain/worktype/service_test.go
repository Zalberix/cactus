package worktype

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestManifestHash проверяет детерминированность хэша манифеста.
func TestManifestHash(t *testing.T) {
	t.Run("same input produces same hash", func(t *testing.T) {
		manifest := json.RawMessage(`{"version":"1.0","settings":{"key":"value"}}`)

		hash1, err := ManifestHash(manifest)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		hash2, err := ManifestHash(manifest)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		if hash1 != hash2 {
			t.Errorf("expected same hash, got %q and %q", hash1, hash2)
		}
	})

	t.Run("different input produces different hash", func(t *testing.T) {
		manifest1 := json.RawMessage(`{"version":"1.0"}`)
		manifest2 := json.RawMessage(`{"version":"2.0"}`)

		hash1, err := ManifestHash(manifest1)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		hash2, err := ManifestHash(manifest2)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		if hash1 == hash2 {
			t.Errorf("expected different hashes but got same: %q", hash1)
		}
	})

	t.Run("hash is 64 char hex string", func(t *testing.T) {
		manifest := json.RawMessage(`{"test":"data"}`)
		hash, err := ManifestHash(manifest)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		if len(hash) != 64 {
			t.Errorf("expected 64-char hash, got len=%d: %q", len(hash), hash)
		}
	})
}

// TestComputeWorkerStatus проверяет вычисление статуса воркера.
func TestComputeWorkerStatus(t *testing.T) {
	t.Run("hasRunningStep=true returns working", func(t *testing.T) {
		// Даже с устаревшим heartbeat — если есть running step → working
		staleHeartbeat := time.Now().Add(-5 * time.Minute)
		status := ComputeWorkerStatus(staleHeartbeat, true)
		if status != WorkerStatusWorking {
			t.Errorf("expected working, got %q", status)
		}
	})

	t.Run("recent heartbeat returns ready", func(t *testing.T) {
		// Heartbeat < 90 секунд назад
		recentHeartbeat := time.Now().Add(-30 * time.Second)
		status := ComputeWorkerStatus(recentHeartbeat, false)
		if status != WorkerStatusReady {
			t.Errorf("expected ready, got %q", status)
		}
	})

	t.Run("stale heartbeat returns offline", func(t *testing.T) {
		// Heartbeat > 90 секунд назад
		staleHeartbeat := time.Now().Add(-2 * time.Minute)
		status := ComputeWorkerStatus(staleHeartbeat, false)
		if status != WorkerStatusOffline {
			t.Errorf("expected offline, got %q", status)
		}
	})
}

// TestGenerateBootstrapToken проверяет формат bootstrap-токена.
func TestGenerateBootstrapToken(t *testing.T) {
	plaintext, hash, err := GenerateBootstrapToken()
	if err != nil {
		t.Fatalf("GenerateBootstrapToken error: %v", err)
	}

	t.Run("plaintext is 64-char hex string", func(t *testing.T) {
		if len(plaintext) != 64 {
			t.Errorf("expected 64-char plaintext, got len=%d: %q", len(plaintext), plaintext)
		}
		if !isHex(plaintext) {
			t.Errorf("expected hex string, got: %q", plaintext)
		}
	})

	t.Run("hash is 64-char hex string", func(t *testing.T) {
		if len(hash) != 64 {
			t.Errorf("expected 64-char hash, got len=%d: %q", len(hash), hash)
		}
		if !isHex(hash) {
			t.Errorf("expected hex string, got: %q", hash)
		}
	})

	t.Run("plaintext and hash are different", func(t *testing.T) {
		if plaintext == hash {
			t.Error("plaintext and hash should be different")
		}
	})

	t.Run("two calls produce different tokens", func(t *testing.T) {
		p2, h2, err := GenerateBootstrapToken()
		if err != nil {
			t.Fatalf("GenerateBootstrapToken error: %v", err)
		}
		if plaintext == p2 {
			t.Error("two calls should produce different plaintexts")
		}
		if hash == h2 {
			t.Error("two calls should produce different hashes")
		}
	})
}

// isHex проверяет что строка состоит из hex-символов.
func isHex(s string) bool {
	s = strings.ToLower(s)
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
