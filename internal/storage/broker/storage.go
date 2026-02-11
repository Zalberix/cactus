package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"cactus/pkg/contracts"
)

type Broker struct {
	client *redis.Client
}

func New(ctx context.Context, options *redis.Options) (*Broker, error) {
	client := redis.NewClient(options)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к брокеру: %w", err)
	}

	return &Broker{client: client}, nil
}

func (b *Broker) Close() error {
	return b.client.Close()
}

func (b *Broker) EnsureStreamGroup(ctx context.Context, streamName, groupName string) error {
	keyType, err := b.client.Type(ctx, streamName).Result()
	if err != nil {
		return fmt.Errorf("не удалось получить тип записи по ключу %s: %w", streamName, err)
	}

	switch keyType {
	case "none":
		err = b.client.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
		if err != nil {
			return fmt.Errorf("не удалось создать группу для сообщений: %w", err)
		}
	case "stream":
		info, err := b.client.XInfoStream(ctx, streamName).Result()
		if err != nil {
			return fmt.Errorf("не удалось получить информацию о стриме: %w", err)
		}
		if info.Groups == 0 {
			err = b.client.XGroupCreate(ctx, streamName, groupName, "0").Err()
			if err != nil {
				return fmt.Errorf("не удалось создать группу для сообщений: %w", err)
			}
		}
	}
	return nil
}

func (b *Broker) AddMessageToQueue(
	ctx context.Context,
	streamName string,
	messageDTO contracts.MessageValueInMessageQueue,
	systemDTO contracts.SystemValueInMessageQueue,
	pipelineDTO contracts.PipelineValueInMessageQueue,
) error {
	messageJSON, err := json.Marshal(messageDTO)
	if err != nil {
		return fmt.Errorf("не удалось сформировать JSON из сообщения: %w", err)
	}

	systemJSON, err := json.Marshal(systemDTO)
	if err != nil {
		return fmt.Errorf("не удалось сформировать JSON из системы: %w", err)
	}

	pipelineJSON, err := json.Marshal(pipelineDTO)
	if err != nil {
		return fmt.Errorf("не удалось сформировать JSON из pipeline: %w", err)
	}

	err = b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"message":  messageJSON,
			"system":   systemJSON,
			"pipeline": pipelineJSON,
		},
	}).Err()
	if err != nil {
		return fmt.Errorf("не удалось добавить сообщение в стрим redis: %w", err)
	}

	return nil
}

func (b *Broker) SendMetaEvent(ctx context.Context, maxPriority int32, endpoint string) error {
	err := b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "event:meta",
		Values: map[string]interface{}{
			"max_priority":      maxPriority,
			"register_endpoint": endpoint,
		},
	}).Err()
	if err != nil {
		return fmt.Errorf("не удалось записать hostname в redis: %w", err)
	}

	return nil
}
