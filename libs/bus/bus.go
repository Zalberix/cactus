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

type Options struct {
	URL             string
	CAFile          string
	CredentialsFile string
}

type StreamOption func(*jetstream.StreamConfig)

func WithAllowDirect() StreamOption {
	return func(cfg *jetstream.StreamConfig) {
		cfg.AllowDirect = true
	}
}

func newStreamConfig(name string, subjects []string, opts ...StreamOption) jetstream.StreamConfig {
	cfg := jetstream.StreamConfig{
		Name:     name,
		Subjects: append([]string(nil), subjects...),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func NewWithOptions(opts Options) (*Bus, error) {
	natsOpts := []nats.Option{}
	if opts.CAFile != "" {
		natsOpts = append(natsOpts, nats.RootCAs(opts.CAFile))
	}
	if opts.CredentialsFile != "" {
		natsOpts = append(natsOpts, nats.UserCredentials(opts.CredentialsFile))
	}
	nc, err := nats.Connect(opts.URL, natsOpts...)
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

func New(url string) (*Bus, error) {
	return NewWithOptions(Options{URL: url})
}

func NewFx(cfg *config.Config) (*Bus, error) {
	return NewWithOptions(Options{
		URL:             cfg.Nats.URL,
		CAFile:          cfg.Nats.CAFile,
		CredentialsFile: cfg.Nats.CredentialsFile,
	})
}

func (b *Bus) Close() {
	b.nc.Drain()
}

// EnsureStream создаёт JetStream-стрим если не существует.
// name — имя стрима (A-Z, 0-9, дефис, подчёркивание).
// subjects — список NATS-субъектов, которые стрим перехватывает (например "messages.>").
func (b *Bus) EnsureStream(ctx context.Context, name string, subjects []string, opts ...StreamOption) error {
	_, err := b.js.CreateOrUpdateStream(ctx, newStreamConfig(name, subjects, opts...))
	if err != nil {
		return fmt.Errorf("не удалось создать/обновить стрим %s: %w", name, err)
	}
	return nil
}

// EnsureStreamWithMaxAge создаёт JetStream-стрим с ограничением по возрасту сообщений.
func (b *Bus) EnsureStreamWithMaxAge(ctx context.Context, name string, subjects []string, maxAge time.Duration) error {
	cfg := newStreamConfig(name, subjects)
	cfg.MaxAge = maxAge
	_, err := b.js.CreateOrUpdateStream(ctx, cfg)
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
func (b *Bus) SubscribeMsg(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	sub, err := b.nc.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("РѕС€РёР±РєР° РїРѕРґРїРёСЃРєРё РЅР° %s: %w", subject, err)
	}
	return sub, nil
}

func (b *Bus) JS() jetstream.JetStream {
	return b.js
}

// SubjectToStreamName конвертирует NATS-субъект в имя стрима:
// "messages.email.smtp.w-1" → "MESSAGES"
func SubjectToStreamName(subject string) string {
	parts := strings.SplitN(subject, ".", 2)
	return strings.ToUpper(parts[0])
}
