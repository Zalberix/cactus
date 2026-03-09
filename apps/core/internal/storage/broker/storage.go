package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zalberix/cactus/libs/bus"
	"github.com/zalberix/cactus/libs/pipeline"
)

type Broker struct {
	bus *bus.Bus
}

func New(b *bus.Bus) *Broker {
	return &Broker{bus: b}
}

// EnsureStreamGroup гарантирует существование JetStream-стрима для данной очереди.
// streamName — субъект в формате "messages.type.kind.w-N".
func (b *Broker) EnsureStreamGroup(ctx context.Context, streamName, _ string) error {
	natsStream := subjectToStreamName(streamName)
	rootSubject := strings.SplitN(streamName, ".", 2)[0] + ".>"
	return b.bus.EnsureStream(ctx, natsStream, []string{rootSubject})
}

// AddMessageToQueue публикует сообщение в JetStream.
func (b *Broker) AddMessageToQueue(
	ctx context.Context,
	subject string,
	messageDTO pipeline.MessageInQueue,
	systemDTO pipeline.SystemInQueue,
	pipelineDTO pipeline.PipelineInQueue,
) error {
	payload := map[string]json.RawMessage{}

	msgJSON, err := json.Marshal(messageDTO)
	if err != nil {
		return fmt.Errorf("ошибка сериализации message: %w", err)
	}
	sysJSON, err := json.Marshal(systemDTO)
	if err != nil {
		return fmt.Errorf("ошибка сериализации system: %w", err)
	}
	plJSON, err := json.Marshal(pipelineDTO)
	if err != nil {
		return fmt.Errorf("ошибка сериализации pipeline: %w", err)
	}

	payload["message"] = msgJSON
	payload["system"] = sysJSON
	payload["pipeline"] = plJSON

	return b.bus.PublishJS(ctx, subject, payload)
}

// SendMetaEvent публикует событие регистрации сервера в NATS.
func (b *Broker) SendMetaEvent(ctx context.Context, maxPriority int32, endpoint string) error {
	return b.bus.PublishJS(ctx, "event.meta", map[string]any{
		"max_priority":      maxPriority,
		"register_endpoint": endpoint,
	})
}

// Close закрывает соединение с NATS.
func (b *Broker) Close() {
	b.bus.Close()
}

// subjectToStreamName: "messages.email.smtp.w-1" → "MESSAGES"
func subjectToStreamName(subject string) string {
	parts := strings.SplitN(subject, ".", 2)
	return strings.ToUpper(parts[0])
}
