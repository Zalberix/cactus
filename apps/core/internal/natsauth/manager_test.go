package natsauth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	natsjwt "github.com/nats-io/jwt/v2"
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

func TestJWTManagerIssueWorkerRejectsSeedThatDoesNotMatchAccountJWT(t *testing.T) {
	seedKP, err := nkeys.CreateAccount()
	if err != nil {
		t.Fatalf("create seed account key: %v", err)
	}
	defer seedKP.Wipe()
	seed, err := seedKP.Seed()
	if err != nil {
		t.Fatalf("read seed account seed: %v", err)
	}

	jwtKP, err := nkeys.CreateAccount()
	if err != nil {
		t.Fatalf("create jwt account key: %v", err)
	}
	defer jwtKP.Wipe()
	jwtPublic, err := jwtKP.PublicKey()
	if err != nil {
		t.Fatalf("read jwt account public key: %v", err)
	}
	accountJWT, err := natsjwt.NewAccountClaims(jwtPublic).Encode(jwtKP)
	if err != nil {
		t.Fatalf("encode account jwt: %v", err)
	}

	dir := t.TempDir()
	seedFile := filepath.Join(dir, "account.seed")
	if err := os.WriteFile(seedFile, seed, 0o600); err != nil {
		t.Fatalf("write account seed: %v", err)
	}
	accountJWTFile := filepath.Join(dir, "account.jwt")
	if err := os.WriteFile(accountJWTFile, []byte(accountJWT), 0o600); err != nil {
		t.Fatalf("write account jwt: %v", err)
	}

	manager := NewJWTManager("", accountJWTFile, "CACTUS_TEST_MISSING_NATS_ACCOUNT_SEED", seedFile, nil)

	_, err = manager.IssueWorker(context.Background(), WorkerScope{
		OrganizationID: 1,
		WorkTypeID:     2,
		WorkerID:       3,
	})
	if err == nil {
		t.Fatalf("expected account seed/account jwt mismatch error")
	}
	if !strings.Contains(err.Error(), "account jwt subject") {
		t.Fatalf("error = %q, want account jwt subject mismatch", err.Error())
	}
}

func TestJWTManagerRevokeWorkerDoesNotSelfSignOperatorAccountJWT(t *testing.T) {
	operatorKP, err := nkeys.CreateOperator()
	if err != nil {
		t.Fatalf("create operator key: %v", err)
	}
	defer operatorKP.Wipe()
	operatorPublicKey, err := operatorKP.PublicKey()
	if err != nil {
		t.Fatalf("read operator public key: %v", err)
	}

	accountKP, err := nkeys.CreateAccount()
	if err != nil {
		t.Fatalf("create account key: %v", err)
	}
	defer accountKP.Wipe()
	accountPublicKey, err := accountKP.PublicKey()
	if err != nil {
		t.Fatalf("read account public key: %v", err)
	}
	accountSeed, err := accountKP.Seed()
	if err != nil {
		t.Fatalf("read account seed: %v", err)
	}

	userKP, err := nkeys.CreateUser()
	if err != nil {
		t.Fatalf("create user key: %v", err)
	}
	defer userKP.Wipe()
	userPublicKey, err := userKP.PublicKey()
	if err != nil {
		t.Fatalf("read user public key: %v", err)
	}

	accountClaims := natsjwt.NewAccountClaims(accountPublicKey)
	accountClaims.Name = "CACTUS"
	accountJWT, err := accountClaims.Encode(operatorKP)
	if err != nil {
		t.Fatalf("encode operator-signed account jwt: %v", err)
	}

	dir := t.TempDir()
	accountSeedFile := filepath.Join(dir, "account.seed")
	if err := os.WriteFile(accountSeedFile, accountSeed, 0o600); err != nil {
		t.Fatalf("write account seed: %v", err)
	}
	accountJWTFile := filepath.Join(dir, "account.jwt")
	if err := os.WriteFile(accountJWTFile, []byte(accountJWT), 0o600); err != nil {
		t.Fatalf("write account jwt: %v", err)
	}

	updater := &recordingAccountClaimsUpdater{}
	manager := NewJWTManager(accountPublicKey, accountJWTFile, "CACTUS_TEST_MISSING_NATS_ACCOUNT_SEED", accountSeedFile, updater)

	err = manager.RevokeWorker(context.Background(), userPublicKey)
	if err == nil {
		t.Fatalf("expected error when operator seed is unavailable")
	}
	if !strings.Contains(err.Error(), "operator seed") {
		t.Fatalf("error = %q, want operator seed error", err.Error())
	}
	if updater.jwt != "" {
		t.Fatalf("updater must not be called when account jwt cannot be signed by %s", operatorPublicKey)
	}

	rawAccountJWT, err := os.ReadFile(accountJWTFile)
	if err != nil {
		t.Fatalf("read account jwt: %v", err)
	}
	if string(rawAccountJWT) != accountJWT {
		t.Fatalf("account jwt was modified without operator seed")
	}
}

func TestJWTManagerRevokeWorkerSignsOperatorAccountJWTWithOperatorSeed(t *testing.T) {
	operatorKP, err := nkeys.CreateOperator()
	if err != nil {
		t.Fatalf("create operator key: %v", err)
	}
	defer operatorKP.Wipe()
	operatorPublicKey, err := operatorKP.PublicKey()
	if err != nil {
		t.Fatalf("read operator public key: %v", err)
	}
	operatorSeed, err := operatorKP.Seed()
	if err != nil {
		t.Fatalf("read operator seed: %v", err)
	}

	accountKP, err := nkeys.CreateAccount()
	if err != nil {
		t.Fatalf("create account key: %v", err)
	}
	defer accountKP.Wipe()
	accountPublicKey, err := accountKP.PublicKey()
	if err != nil {
		t.Fatalf("read account public key: %v", err)
	}
	accountSeed, err := accountKP.Seed()
	if err != nil {
		t.Fatalf("read account seed: %v", err)
	}

	userKP, err := nkeys.CreateUser()
	if err != nil {
		t.Fatalf("create user key: %v", err)
	}
	defer userKP.Wipe()
	userPublicKey, err := userKP.PublicKey()
	if err != nil {
		t.Fatalf("read user public key: %v", err)
	}

	accountClaims := natsjwt.NewAccountClaims(accountPublicKey)
	accountClaims.Name = "CACTUS"
	accountJWT, err := accountClaims.Encode(operatorKP)
	if err != nil {
		t.Fatalf("encode operator-signed account jwt: %v", err)
	}

	dir := t.TempDir()
	accountSeedFile := filepath.Join(dir, "account.seed")
	if err := os.WriteFile(accountSeedFile, accountSeed, 0o600); err != nil {
		t.Fatalf("write account seed: %v", err)
	}
	operatorSeedFile := filepath.Join(dir, "operator.seed")
	if err := os.WriteFile(operatorSeedFile, operatorSeed, 0o600); err != nil {
		t.Fatalf("write operator seed: %v", err)
	}
	accountJWTFile := filepath.Join(dir, "account.jwt")
	if err := os.WriteFile(accountJWTFile, []byte(accountJWT), 0o600); err != nil {
		t.Fatalf("write account jwt: %v", err)
	}

	updater := &recordingAccountClaimsUpdater{}
	manager := NewJWTManager(
		accountPublicKey,
		accountJWTFile,
		"CACTUS_TEST_MISSING_NATS_ACCOUNT_SEED",
		accountSeedFile,
		updater,
		WithOperatorSeedFile(operatorSeedFile),
	)

	if err := manager.RevokeWorker(context.Background(), userPublicKey); err != nil {
		t.Fatalf("revoke worker: %v", err)
	}
	if updater.jwt == "" {
		t.Fatalf("expected updater to receive account jwt")
	}

	updatedClaims, err := natsjwt.DecodeAccountClaims(updater.jwt)
	if err != nil {
		t.Fatalf("decode updated account jwt: %v", err)
	}
	if updatedClaims.Issuer != operatorPublicKey {
		t.Fatalf("updated issuer = %q, want %q", updatedClaims.Issuer, operatorPublicKey)
	}
	if updatedClaims.Subject != accountPublicKey {
		t.Fatalf("updated subject = %q, want %q", updatedClaims.Subject, accountPublicKey)
	}
	if _, ok := updatedClaims.Revocations[userPublicKey]; !ok {
		t.Fatalf("expected revocation for %s", userPublicKey)
	}

	rawAccountJWT, err := os.ReadFile(accountJWTFile)
	if err != nil {
		t.Fatalf("read account jwt: %v", err)
	}
	if string(rawAccountJWT) != updater.jwt {
		t.Fatalf("account jwt file does not match pushed jwt")
	}
}

type recordingAccountClaimsUpdater struct {
	jwt string
}

func (u *recordingAccountClaimsUpdater) UpdateAccountJWT(_ context.Context, accountJWT string) error {
	u.jwt = accountJWT
	return nil
}
