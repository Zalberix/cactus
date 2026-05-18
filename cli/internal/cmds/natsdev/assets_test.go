package natsdev

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAssetsCreatesLocalNATSFiles(t *testing.T) {
	root := t.TempDir()
	result, err := GenerateAssets(root, false)
	if err != nil {
		t.Fatalf("GenerateAssets: %v", err)
	}
	if result.AccountPublicKey == "" {
		t.Fatalf("expected account public key")
	}
	if result.AccountSeed == "" {
		t.Fatalf("expected account seed")
	}

	files := []string{
		"configs/nats/certs/ca.pem",
		"configs/nats/certs/server.pem",
		"configs/nats/certs/server-key.pem",
		"configs/nats/jwt/operator.jwt",
		"configs/nats/jwt/account.jwt",
		"configs/nats/jwt/account.seed",
		"configs/nats/creds/core.creds",
		"configs/nats/creds/sys.creds",
		"configs/nats/dev.env",
		filepath.Join("docker/assets/nats-jwt", result.AccountPublicKey+".jwt"),
	}
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(root, file)); err != nil {
			t.Fatalf("expected %s: %v", file, err)
		}
	}
}
