package natsauth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	natsjwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type AccountClaimsUpdater interface {
	UpdateAccountJWT(ctx context.Context, accountJWT string) error
}

type JWTManager struct {
	accountPublicKey string
	accountJWTFile   string
	accountSeedEnv   string
	accountSeedFile  string
	updater          AccountClaimsUpdater
}

func NewJWTManager(accountPublicKey, accountJWTFile, accountSeedEnv, accountSeedFile string, updater AccountClaimsUpdater) *JWTManager {
	return &JWTManager{
		accountPublicKey: accountPublicKey,
		accountJWTFile:   accountJWTFile,
		accountSeedEnv:   accountSeedEnv,
		accountSeedFile:  accountSeedFile,
		updater:          updater,
	}
}

func (m *JWTManager) IssueWorker(_ context.Context, scope WorkerScope) (WorkerCredentials, error) {
	accountSeed, err := m.loadAccountSeed()
	if err != nil {
		return WorkerCredentials{}, err
	}
	accountKP, err := nkeys.FromSeed([]byte(accountSeed))
	if err != nil {
		return WorkerCredentials{}, fmt.Errorf("load account signing seed: %w", err)
	}
	defer accountKP.Wipe()
	accountPublicKey := m.accountPublicKey
	if accountPublicKey == "" {
		accountPublicKey, err = accountKP.PublicKey()
		if err != nil {
			return WorkerCredentials{}, fmt.Errorf("read account public key: %w", err)
		}
	}

	userKP, err := nkeys.CreateUser()
	if err != nil {
		return WorkerCredentials{}, fmt.Errorf("create nats user key: %w", err)
	}
	defer userKP.Wipe()

	userPublicKey, err := userKP.PublicKey()
	if err != nil {
		return WorkerCredentials{}, fmt.Errorf("read user public key: %w", err)
	}
	userSeed, err := userKP.Seed()
	if err != nil {
		return WorkerCredentials{}, fmt.Errorf("read user seed: %w", err)
	}

	perms := WorkerPermissions(scope)
	claims := natsjwt.NewUserClaims(userPublicKey)
	claims.Name = fmt.Sprintf("worker-%d-org-%d-worktype-%d", scope.WorkerID, scope.OrganizationID, scope.WorkTypeID)
	claims.Permissions.Pub.Allow = natsjwt.StringList(perms.PublishAllow)
	claims.Permissions.Sub.Allow = natsjwt.StringList(perms.SubscribeAllow)

	userJWT, err := claims.Encode(accountKP)
	if err != nil {
		return WorkerCredentials{}, fmt.Errorf("encode worker user jwt: %w", err)
	}
	creds, err := natsjwt.FormatUserConfig(userJWT, userSeed)
	if err != nil {
		return WorkerCredentials{}, fmt.Errorf("format worker creds: %w", err)
	}

	return WorkerCredentials{
		AccountPublicKey: accountPublicKey,
		UserPublicKey:    userPublicKey,
		UserJWT:          userJWT,
		UserSeed:         string(userSeed),
		Creds:            creds,
		Permissions:      perms,
	}, nil
}

func (m *JWTManager) RevokeWorker(ctx context.Context, userPublicKey string) error {
	if userPublicKey == "" {
		return fmt.Errorf("user public key is required")
	}
	accountSeed, err := m.loadAccountSeed()
	if err != nil {
		return err
	}
	accountKP, err := nkeys.FromSeed([]byte(accountSeed))
	if err != nil {
		return fmt.Errorf("load account seed: %w", err)
	}
	defer accountKP.Wipe()

	rawJWT, err := os.ReadFile(m.accountJWTFile)
	if err != nil {
		return fmt.Errorf("read account jwt: %w", err)
	}
	claims, err := natsjwt.DecodeAccountClaims(strings.TrimSpace(string(rawJWT)))
	if err != nil {
		return fmt.Errorf("decode account jwt: %w", err)
	}
	claims.RevokeAt(userPublicKey, time.Now())
	updatedJWT, err := claims.Encode(accountKP)
	if err != nil {
		return fmt.Errorf("encode revoked account jwt: %w", err)
	}
	if err := os.WriteFile(m.accountJWTFile, []byte(updatedJWT), 0o600); err != nil {
		return fmt.Errorf("write account jwt: %w", err)
	}
	if m.updater == nil {
		return fmt.Errorf("account claims updater is required")
	}
	if err := m.updater.UpdateAccountJWT(ctx, updatedJWT); err != nil {
		return fmt.Errorf("push account jwt update: %w", err)
	}
	return nil
}

func (m *JWTManager) loadAccountSeed() (string, error) {
	if m.accountSeedEnv != "" {
		if accountSeed := strings.TrimSpace(os.Getenv(m.accountSeedEnv)); accountSeed != "" {
			return accountSeed, nil
		}
	}
	if m.accountSeedFile != "" {
		raw, err := os.ReadFile(m.accountSeedFile)
		if err != nil {
			return "", fmt.Errorf("read account seed file %q: %w", m.accountSeedFile, err)
		}
		if accountSeed := strings.TrimSpace(string(raw)); accountSeed != "" {
			return accountSeed, nil
		}
	}
	if m.accountSeedEnv != "" && m.accountSeedFile != "" {
		return "", fmt.Errorf("%s or account seed file %q is required", m.accountSeedEnv, m.accountSeedFile)
	}
	if m.accountSeedEnv != "" {
		return "", fmt.Errorf("%s is required", m.accountSeedEnv)
	}
	return "", fmt.Errorf("account seed is required")
}
