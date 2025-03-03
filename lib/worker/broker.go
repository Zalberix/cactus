package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BrokerRedis struct {
	rdb *redis.Client
}

type (
	BrokerMessages map[string][]BrokerMessage
	BrokerMessage  struct {
		ID     string
		Values map[string]interface{}
	}
)

func NewBrokerRedis(rdb *redis.Client) *BrokerRedis {
	return &BrokerRedis{
		rdb: rdb,
	}
}

func (broker *BrokerRedis) Ack(ctx context.Context, name string, groupName string, id string) error {
	err := broker.rdb.XAck(ctx, name, groupName, id).Err()
	return err
}

func (broker *BrokerRedis) Read(ctx context.Context, streams []string, block time.Duration, count int64) (BrokerMessages, error) {
	args := &redis.XReadArgs{
		Streams: streams,
		Block:   block,
		Count:   count,
	}

	msgs, err := broker.rdb.XRead(ctx, args).Result()
	if err != nil {
		return nil, err
	}

	messages := make(BrokerMessages)

	for _, msg := range msgs {
		for _, m := range msg.Messages {
			messages[msg.Stream] = append(messages[msg.Stream], BrokerMessage{
				ID:     m.ID,
				Values: m.Values,
			})
		}
	}

	return messages, nil
}

func (broker *BrokerRedis) ReadGroup(ctx context.Context, group string, consumer string, streams []string, block time.Duration, count int64) (BrokerMessages, error) {
	args := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  streams,
		Block:    block,
		Count:    count,
	}

	msgs, err := broker.rdb.XReadGroup(ctx, args).Result()
	if err != nil {
		return nil, err
	}

	messages := make(BrokerMessages)

	for _, msg := range msgs {
		for _, m := range msg.Messages {
			messages[msg.Stream] = append(messages[msg.Stream], BrokerMessage{
				ID:     m.ID,
				Values: m.Values,
			})
		}
	}

	return messages, nil
}

func (broker *BrokerRedis) Add(ctx context.Context, stream string, id string, values map[string]interface{}) error {
	return broker.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "event:pipeline",
		ID:     "*",
		Values: values,
	}).Err()
}

func (broker *BrokerRedis) ReadLatestMessages(ctx context.Context, stream string, count int64) ([]BrokerMessage, error) {
	msgs, err := broker.rdb.XRevRangeN(ctx, stream, "+", "-", count).Result()
	if err != nil {
		return []BrokerMessage{}, fmt.Errorf("ошибка чтения из стрима: %w", err)
	}

	messages := make([]BrokerMessage, 0, len(msgs))

	for _, m := range msgs {
		messages = append(messages, BrokerMessage{
			ID:     m.ID,
			Values: m.Values,
		})
	}

	return messages, nil
}

// TODO пересмотреть удаление стримов, так как один не правильно написанный worker может удалять
// стримы что не верно. Скорее лучше говорить что запустить не возможно так как ключи под стримы уже заняты
func (broker *BrokerRedis) CreateStreamIfNotExist(ctx context.Context, key string, group string) error {
	typeRes, err := broker.rdb.Type(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("ошибка получения типа ключа %s: %w", key, err)
	}

	switch typeRes {
	case "none":
		if group != "" {
			err = broker.rdb.XGroupCreateMkStream(ctx, key, group, "$").Err()
			if err != nil {
				return fmt.Errorf("ошибка создания группы %s в стриме %s: %w", group, key, err)
			}
		} else {
			_, err = broker.rdb.XAdd(ctx, &redis.XAddArgs{
				Stream: key,
				Values: map[string]interface{}{"init": "stream"},
			}).Result()
			if err != nil {
				return fmt.Errorf("ошибка создания стрима %s: %w", key, err)
			}
		}
	case "stream":
		info, err := broker.rdb.XInfoStream(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("ошибка получения информации о стриме %s: %w", key, err)
		}

		if group != "" && info.Groups == 0 {
			return fmt.Errorf(
				"запуск чтения стрима не возможен так как стрим с именем %v без группы, а для работы нужен стрим с группой",
				key,
			)
		} else if group == "" && info.Groups > 0 {
			return fmt.Errorf(
				"запуск чтения стрима не возможен так как стрим с именем %v с группой, а для работы нужен стрим без группы",
				key,
			)
		}
	default:
		err = broker.rdb.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("ошибка удаления ключа %s: %w", key, err)
		}
		return broker.CreateStreamIfNotExist(ctx, key, group)
	}

	return nil
}
