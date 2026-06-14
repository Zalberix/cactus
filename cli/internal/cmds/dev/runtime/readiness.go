package runtime

import (
	"context"
	"fmt"
	"net"
	"time"
)

func ImmediateReady() ReadyProbe {
	return func(context.Context) error {
		return nil
	}
}

func PortReadyProbe(port int, interval time.Duration, timeout time.Duration) ReadyProbe {
	if interval <= 0 {
		interval = 300 * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err == nil {
				_ = conn.Close()
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
	}
}
