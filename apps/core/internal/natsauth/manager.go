package natsauth

import "context"

type WorkerScope struct {
	OrganizationID int32
	WorkTypeID     int32
	WorkerID       int32
}

type Permissions struct {
	PublishAllow   []string `json:"publish_allow"`
	SubscribeAllow []string `json:"subscribe_allow"`
}

type WorkerCredentials struct {
	AccountPublicKey string
	UserPublicKey    string
	UserJWT          string
	UserSeed         string
	Creds            []byte
	Permissions      Permissions
}

type Manager interface {
	AccountPublicKey(ctx context.Context) (string, error)
	IssueWorker(ctx context.Context, scope WorkerScope) (WorkerCredentials, error)
	RevokeWorker(ctx context.Context, userPublicKey string) error
}
