package natsauth

import (
	"context"
	"fmt"
)

type NoopManager struct{}

func (NoopManager) IssueWorker(_ context.Context, scope WorkerScope) (WorkerCredentials, error) {
	perms := WorkerPermissions(scope)
	return WorkerCredentials{
		AccountPublicKey: "ADUMMY",
		UserPublicKey:    fmt.Sprintf("UDUMMY%d", scope.WorkerID),
		UserJWT:          "dummy.jwt",
		UserSeed:         "SUdummyseed",
		Creds:            []byte("dummy.creds"),
		Permissions:      perms,
	}, nil
}

func (NoopManager) RevokeWorker(_ context.Context, _ string) error {
	return nil
}
