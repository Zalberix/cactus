package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

const defaultWorkerTimeout = 5 * time.Minute

// waitForResult ожидает результат от воркера через NATS JetStream ephemeral consumer.
// Подписывается на replySubject в RESULTS stream с timeout.
// Per D-05: timeout берётся из StepDef.Timeout (заполняется из WorkerSettingsRevision),
// если 0 — используется defaultWorkerTimeout (5 min).
func waitForResult(ctx context.Context, js jetstream.JetStream, replySubject string, timeout time.Duration) (temporaltypes.WorkerResult, error) {
	if timeout <= 0 {
		timeout = defaultWorkerTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Ephemeral consumer — без имени, автоудаление (per RESEARCH Pitfall 3)
	cons, err := js.CreateConsumer(ctx, "RESULTS", jetstream.ConsumerConfig{
		FilterSubject:     replySubject,
		AckPolicy:         jetstream.AckExplicitPolicy,
		InactiveThreshold: 30 * time.Second,
	})
	if err != nil {
		return temporaltypes.WorkerResult{}, fmt.Errorf("create result consumer for %s: %w", replySubject, err)
	}

	// Fetch one message with context timeout
	msgs, err := cons.Fetch(1, jetstream.FetchMaxWait(timeout))
	if err != nil {
		return temporaltypes.WorkerResult{}, fmt.Errorf("fetch result from %s: %w", replySubject, err)
	}

	for msg := range msgs.Messages() {
		var result temporaltypes.WorkerResult
		if err := json.Unmarshal(msg.Data(), &result); err != nil {
			_ = msg.Nak()
			return temporaltypes.WorkerResult{}, fmt.Errorf("unmarshal worker result: %w", err)
		}
		_ = msg.Ack()
		return result, nil
	}

	if msgs.Error() != nil {
		return temporaltypes.WorkerResult{}, fmt.Errorf("result fetch error on %s: %w", replySubject, msgs.Error())
	}

	return temporaltypes.WorkerResult{}, fmt.Errorf("timeout waiting for worker result on %s after %v", replySubject, timeout)
}
