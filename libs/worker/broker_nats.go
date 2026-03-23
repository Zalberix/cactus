package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ErrNoMessages возвращается когда нет доступных сообщений.
var ErrNoMessages = errors.New("нет доступных сообщений")

type (
	BrokerMessages map[string][]BrokerMessage
	BrokerMessage  struct {
		ID     string
		Values map[string]interface{}
	}
)

// BrokerNATS реализует интерфейс Broker через NATS JetStream.
type BrokerNATS struct {
	nc *nats.Conn
	js jetstream.JetStream

	mu        sync.Mutex
	pending   map[string]jetstream.Msg
	consumers map[string]jetstream.Consumer
	subs      map[string]*nats.Subscription
	subChans  map[string]chan *nats.Msg
}

func NewBrokerNATS(url string) (*BrokerNATS, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("не удалось инициализировать JetStream: %w", err)
	}

	return &BrokerNATS{
		nc:        nc,
		js:        js,
		pending:   make(map[string]jetstream.Msg),
		consumers: make(map[string]jetstream.Consumer),
		subs:      make(map[string]*nats.Subscription),
		subChans:  make(map[string]chan *nats.Msg),
	}, nil
}

func (b *BrokerNATS) Close() {
	b.mu.Lock()
	for _, sub := range b.subs {
		sub.Unsubscribe()
	}
	b.mu.Unlock()
	b.nc.Drain()
}

// subjectToStreamName конвертирует "messages.email.smtp.w-1" → "MESSAGES", "event.meta" → "EVENTS".
func subjectToStreamName(subject string) string {
	parts := strings.SplitN(subject, ".", 2)
	return strings.ToUpper(parts[0])
}

func (b *BrokerNATS) Ack(ctx context.Context, name string, groupName string, id string) error {
	b.mu.Lock()
	msg, ok := b.pending[id]
	if ok {
		delete(b.pending, id)
	}
	b.mu.Unlock()

	if !ok {
		return fmt.Errorf("сообщение с ID %s не найдено для подтверждения", id)
	}

	return msg.Ack()
}

// Read подписывается на NATS subject и ждёт сообщения. Используется для event-стримов.
func (b *BrokerNATS) Read(
	ctx context.Context,
	streams []string,
	block time.Duration,
	count int64,
) (BrokerMessages, error) {
	if len(streams) < 2 {
		return nil, fmt.Errorf("streams должен содержать [subject, startID]")
	}
	subject := streams[0]

	b.mu.Lock()
	ch, exists := b.subChans[subject]
	if !exists {
		ch = make(chan *nats.Msg, 64)
		sub, err := b.nc.ChanSubscribe(subject, ch)
		if err != nil {
			b.mu.Unlock()
			return nil, fmt.Errorf("ошибка подписки на %s: %w", subject, err)
		}
		b.subs[subject] = sub
		b.subChans[subject] = ch
	}
	b.mu.Unlock()

	timeout := block
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	messages := make(BrokerMessages)
	var collected []BrokerMessage

	for i := int64(0); i < count; i++ {
		select {
		case msg := <-ch:
			values, err := decodeNATSPayload(msg.Data)
			if err != nil {
				continue
			}
			collected = append(collected, BrokerMessage{
				ID:     fmt.Sprintf("nats-%d", time.Now().UnixNano()),
				Values: values,
			})
		case <-timer.C:
			if len(collected) > 0 {
				messages[subject] = collected
			}
			return messages, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if len(collected) > 0 {
		messages[subject] = collected
	}
	return messages, nil
}

// ReadGroup читает из JetStream pull consumer. Используется для очередей сообщений.
func (b *BrokerNATS) ReadGroup(
	ctx context.Context,
	group string, consumer string,
	streams []string,
	block time.Duration,
	count int64,
) (BrokerMessages, error) {
	if len(streams) < 2 {
		return nil, fmt.Errorf("streams должен содержать [subject, startID]")
	}
	subject := streams[0]

	consumerKey := subject + ":" + group
	b.mu.Lock()
	cons, ok := b.consumers[consumerKey]
	b.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("консьюмер для %s:%s не найден, вызовите CreateStreamIfNotExist", subject, group)
	}

	timeout := block
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	batch, err := cons.Fetch(int(count), jetstream.FetchMaxWait(timeout))
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}

	messages := make(BrokerMessages)
	var collected []BrokerMessage

	for msg := range batch.Messages() {
		values, err := decodeNATSPayload(msg.Data())
		if err != nil {
			msg.Nak()
			continue
		}

		meta, _ := msg.Metadata()
		msgID := fmt.Sprintf("%s-%d", subject, time.Now().UnixNano())
		if meta != nil {
			msgID = fmt.Sprintf("%s-%d", subject, meta.Sequence.Stream)
		}

		b.mu.Lock()
		b.pending[msgID] = msg
		b.mu.Unlock()

		collected = append(collected, BrokerMessage{
			ID:     msgID,
			Values: values,
		})
	}

	if err := batch.Error(); err != nil && !errors.Is(err, jetstream.ErrNoMessages) {
		if len(collected) == 0 {
			return nil, ErrNoMessages
		}
	}

	if len(collected) > 0 {
		messages[subject] = collected
	} else {
		return nil, ErrNoMessages
	}

	return messages, nil
}

// Add публикует сообщение в NATS JetStream.
func (b *BrokerNATS) Add(
	ctx context.Context,
	stream string, id string, values map[string]interface{},
) error {
	payload, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	if _, err := b.js.Publish(ctx, stream, payload); err != nil {
		return fmt.Errorf("ошибка публикации в %s: %w", stream, err)
	}
	return nil
}

// ReadLatestMessages возвращает последние count сообщений из JetStream стрима.
func (b *BrokerNATS) ReadLatestMessages(
	ctx context.Context,
	stream string,
	count int64,
) ([]BrokerMessage, error) {
	streamName := subjectToStreamName(stream)

	s, err := b.js.Stream(ctx, streamName)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения стрима %s: %w", streamName, err)
	}

	rawMsg, err := s.GetLastMsgForSubject(ctx, stream)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения последнего сообщения из %s: %w", stream, err)
	}

	values, err := decodeNATSPayload(rawMsg.Data)
	if err != nil {
		return nil, err
	}

	return []BrokerMessage{{
		ID:     fmt.Sprintf("%d", rawMsg.Sequence),
		Values: values,
	}}, nil
}

// CreateStreamIfNotExist создаёт JetStream стрим и, при необходимости, durable consumer.
func (b *BrokerNATS) CreateStreamIfNotExist(ctx context.Context, key string, group string) error {
	streamName := subjectToStreamName(key)
	wildcard := strings.SplitN(key, ".", 2)[0] + ".>"

	b.mu.Lock()
	defer b.mu.Unlock()

	_, err := b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{wildcard},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания стрима %s: %w", streamName, err)
	}

	if group != "" {
		consumerKey := key + ":" + group
		if _, ok := b.consumers[consumerKey]; !ok {
			consumerName := group + "-" + strings.ReplaceAll(key, ".", "-")
			cons, err := b.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
				Name:          consumerName,
				Durable:       consumerName,
				FilterSubject: key,
				AckPolicy:     jetstream.AckExplicitPolicy,
				AckWait:       30 * time.Second,
			})
			if err != nil {
				return fmt.Errorf("ошибка создания консьюмера %s: %w", consumerName, err)
			}
			b.consumers[consumerKey] = cons
		}
	}

	return nil
}

// decodeNATSPayload декодирует JSON payload NATS сообщения в map[string]interface{}.
// Для совместимости с Redis форматом: nested JSON объекты конвертируются в JSON строки,
// а примитивы (числа) конвертируются в строковое представление.
// Это нужно потому что Redis Streams хранят все значения как строки.
func decodeNATSPayload(data []byte) (map[string]interface{}, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("ошибка десериализации сообщения: %w", err)
	}

	values := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		// Пробуем декодировать как строку
		var s string
		if json.Unmarshal(v, &s) == nil {
			values[k] = s
			continue
		}

		// Для всех остальных типов (числа, объекты, массивы) — сохраняем как JSON строку.
		// Числа: max_priority: 3 → "3"
		// Объекты: pipeline: {"step":1} → "{\"step\":1}"
		values[k] = string(v)
	}

	return values, nil
}
