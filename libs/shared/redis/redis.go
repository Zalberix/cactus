package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, options *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(options)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("неудалость подключиться к redis: %w", err)
	}

	return client, nil
}
