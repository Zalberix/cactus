package natsauth

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	claimsUpdateSubject       = "$SYS.REQ.CLAIMS.UPDATE"
	resolverNATSReconnectWait = time.Second
)

type ResolverUpdater struct {
	nc *nats.Conn
}

func NewResolverUpdater(url, caFile, credentialsFile string) (*ResolverUpdater, error) {
	nc, err := nats.Connect(url, resolverNATSOptions(caFile, credentialsFile)...)
	if err != nil {
		return nil, fmt.Errorf("connect system nats account: %w", err)
	}
	return &ResolverUpdater{nc: nc}, nil
}

func resolverNATSOptions(caFile, credentialsFile string) []nats.Option {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(resolverNATSReconnectWait),
	}
	if caFile != "" {
		opts = append(opts, nats.RootCAs(caFile))
	}
	if credentialsFile != "" {
		opts = append(opts, nats.UserCredentials(credentialsFile))
	}
	return opts
}

func (u *ResolverUpdater) UpdateAccountJWT(ctx context.Context, accountJWT string) error {
	if u == nil || u.nc == nil {
		return fmt.Errorf("resolver updater is not connected")
	}
	msg, err := u.nc.RequestWithContext(ctx, claimsUpdateSubject, []byte(accountJWT))
	if err != nil {
		return fmt.Errorf("claims update request: %w", err)
	}
	if len(msg.Data) == 0 {
		return fmt.Errorf("empty claims update response")
	}
	return nil
}

func (u *ResolverUpdater) Close() error {
	if u == nil || u.nc == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		u.nc.Drain()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
		u.nc.Close()
		return nil
	}
}
