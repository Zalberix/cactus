package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/zalberix/cactus/apps/core/config"
)

type Bus struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func New(url string) (*Bus, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("не удалось инициализировать JetStream: %w", err)
	}

	return &Bus{nc: nc, js: js}, nil
}

func NewFx(cfg *config.Config) (*Bus, error) {
	return New(cfg.Nats.URL)
}

func (b *Bus) Close() {
	b.nc.Drain()
}

// EnsureStream создаёт JetStream-стрим если не существует.
// name — имя стрима (A-Z, 0-9, дефис, подчёркивание).
// subjects — список NATS-субъектов, которые стрим перехватывает (например "messages.>").
func (b *Bus) EnsureStream(ctx context.Context, name string, subjects []string) error {
	_, err := b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: subjects,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать/обновить стрим %s: %w", name, err)
	}
	return nil
}

// EnsureStreamWithMaxAge создаёт JetStream-стрим с ограничением по возрасту сообщений.
func (b *Bus) EnsureStreamWithMaxAge(ctx context.Context, name string, subjects []string, maxAge time.Duration) error {
	_, err := b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: subjects,
		MaxAge:   maxAge,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать/обновить стрим %s: %w", name, err)
	}
	return nil
}

// PublishJS публикует JSON-сообщение в JetStream (персистентная очередь).
func (b *Bus) PublishJS(ctx context.Context, subject string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}
	if _, err = b.js.Publish(ctx, subject, payload); err != nil {
		return fmt.Errorf("ошибка публикации в JetStream (%s): %w", subject, err)
	}
	return nil
}

// Publish публикует JSON-сообщение в core NATS (без персистентности).
func (b *Bus) Publish(subject string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}
	if err = b.nc.Publish(subject, payload); err != nil {
		return fmt.Errorf("ошибка публикации в NATS (%s): %w", subject, err)
	}
	return nil
}

// Subscribe подписывается на core NATS-субъект.
// handler получает сырые байты сообщения.
func (b *Bus) Subscribe(subject string, handler func(data []byte)) (*nats.Subscription, error) {
	sub, err := b.nc.Subscribe(subject, func(m *nats.Msg) {
		handler(m.Data)
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подписки на %s: %w", subject, err)
	}
	return sub, nil
}

// JS возвращает JetStream интерфейс для прямого доступа (consumers, streams).
func (b *Bus) JS() jetstream.JetStream {
	return b.js
}

// SubjectToStreamName конвертирует NATS-субъект в имя стрима:
// "messages.email.smtp.w-1" → "MESSAGES"
func SubjectToStreamName(subject string) string {
	parts := strings.SplitN(subject, ".", 2)
	return strings.ToUpper(parts[0])
}
