package natsauth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nats-io/nkeys"
)

func TestWorkerPermissions(t *testing.T) {
	perms := WorkerPermissions(WorkerScope{OrganizationID: 12, WorkTypeID: 3, WorkerID: 42})

	wantPub := []string{
		"result.org.12.>",
		"config.request.org.12.work_type.3.>",
		"$JS.API.STREAM.INFO.CONFIGS",
		"$JS.API.DIRECT.GET.CONFIGS.config.org.12.work_type.3.>",
		"$JS.API.CONSUMER.CREATE.TASKS.worker-42.task.org.12.work_type.3.>",
		"$JS.API.CONSUMER.MSG.NEXT.TASKS.worker-42",
		"$JS.ACK.TASKS.worker-42.>",
		"_INBOX.>",
	}
	wantSub := []string{"task.org.12.work_type.3.>", "config.org.12.work_type.3.>", "_INBOX.>"}

	if !sameStrings(perms.PublishAllow, wantPub) {
		t.Fatalf("publish permissions mismatch: %#v", perms.PublishAllow)
	}
	if !sameStrings(perms.SubscribeAllow, wantSub) {
		t.Fatalf("subscribe permissions mismatch: %#v", perms.SubscribeAllow)
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestJWTManagerIssueWorkerReadsSeedFileAndDerivesAccountPublicKey(t *testing.T) {
	accountKP, err := nkeys.CreateAccount()
	if err != nil {
		t.Fatalf("create account key: %v", err)
	}
	defer accountKP.Wipe()
	accountSeed, err := accountKP.Seed()
	if err != nil {
		t.Fatalf("read account seed: %v", err)
	}
	accountPublicKey, err := accountKP.PublicKey()
	if err != nil {
		t.Fatalf("read account public key: %v", err)
	}

	seedFile := filepath.Join(t.TempDir(), "account.seed")
	if err := os.WriteFile(seedFile, accountSeed, 0o600); err != nil {
		t.Fatalf("write seed file: %v", err)
	}

	manager := NewJWTManager("", "", "CACTUS_TEST_MISSING_NATS_ACCOUNT_SEED", seedFile, nil)
	creds, err := manager.IssueWorker(context.Background(), WorkerScope{
		OrganizationID: 12,
		WorkTypeID:     3,
		WorkerID:       42,
	})
	if err != nil {
		t.Fatalf("issue worker: %v", err)
	}
	if creds.AccountPublicKey != accountPublicKey {
		t.Fatalf("AccountPublicKey = %q, want %q", creds.AccountPublicKey, accountPublicKey)
	}
	if creds.UserJWT == "" || creds.UserSeed == "" || len(creds.Creds) == 0 {
		t.Fatalf("expected user jwt, seed and creds to be populated")
	}
}
