package wshub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/libs/bus"
	"github.com/zalberix/cactus/libs/pipeline"
)

type PipelineClient struct {
	Hub     *PipelineHub
	Conn    *websocket.Conn
	UUID    string
	message chan []byte
}

func (c *PipelineClient) Listen() {
	for message := range c.message {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func (c *PipelineClient) Write(message []byte) {
	select {
	case c.message <- message:
	default:
		c.Hub.Unregister <- c
	}
}

type ServicePipelineHub interface {
	UpdateStatusPipeline(
		ctx context.Context,
		uuidMessage string,
		step int32,
		status pipeline.Status,
		workerUUID string,
	) (dto.Pipeline, error)
}

type PipelineHub struct {
	clients       map[*PipelineClient]bool
	subscriptions map[string]map[*PipelineClient]bool

	upgrader *websocket.Upgrader
	bus      *bus.Bus
	ctx      context.Context

	Service    ServicePipelineHub
	Register   chan *PipelineClient
	Unregister chan *PipelineClient

	mu sync.Mutex
}

func NewPipelineHub(ctx context.Context, b *bus.Bus, service ServicePipelineHub) *PipelineHub {
	return &PipelineHub{
		clients:       make(map[*PipelineClient]bool),
		subscriptions: make(map[string]map[*PipelineClient]bool),
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(_ *http.Request) bool { return true },
		},
		bus:        b,
		ctx:        ctx,
		Service:    service,
		Register:   make(chan *PipelineClient),
		Unregister: make(chan *PipelineClient),
	}
}

func (hub *PipelineHub) NewClient(ctx context.Context, uuid string, conn *websocket.Conn) *PipelineClient {
	client := &PipelineClient{
		Hub:     hub,
		UUID:    uuid,
		Conn:    conn,
		message: make(chan []byte),
	}
	hub.Register <- client
	return client
}

func (hub *PipelineHub) Upgrader() *websocket.Upgrader {
	return hub.upgrader
}

func (hub *PipelineHub) Send(pm pipeline.PipelineMessage) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	newPipeline, err := hub.Service.UpdateStatusPipeline(
		hub.ctx,
		pm.UUID,
		pm.Step,
		pm.Status,
		pm.WorkeUUID,
	)
	if err != nil {
		slog.Error("Переход к шагу не возможен",
			slog.Any("message", pm),
			slog.Any("err", err.Error()),
		)
		return
	}

	pipelineJSON, err := json.Marshal(newPipeline)
	if err != nil {
		return
	}

	if subs, exists := hub.subscriptions[pm.UUID]; exists {
		for client := range subs {
			select {
			case client.message <- pipelineJSON:
			default:
				close(client.message)
				delete(hub.clients, client)
				delete(subs, client)
			}
		}
		if len(subs) == 0 {
			delete(hub.subscriptions, pm.UUID)
		}
	}
}

// Run запускает обработчик событий и подписку на NATS.
func (hub *PipelineHub) Run() {
	if _, err := hub.bus.Subscribe("event.pipeline", func(data []byte) {
		var pm pipeline.PipelineMessage
		if err := json.Unmarshal(data, &pm); err != nil {
			slog.Error("Сообщение от воркера невозможно обработать",
				slog.String("err", err.Error()),
			)
			return
		}
		hub.Send(pm)
	}); err != nil {
		slog.Error("Ошибка подписки на event.pipeline", slog.String("err", fmt.Sprintf("%v", err)))
		return
	}

	for {
		select {
		case client := <-hub.Register:
			hub.mu.Lock()
			hub.clients[client] = true
			if hub.subscriptions[client.UUID] == nil {
				hub.subscriptions[client.UUID] = make(map[*PipelineClient]bool)
			}
			hub.subscriptions[client.UUID][client] = true
			hub.mu.Unlock()

		case client := <-hub.Unregister:
			hub.mu.Lock()
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				if subs, exists := hub.subscriptions[client.UUID]; exists {
					delete(subs, client)
					if len(subs) == 0 {
						delete(hub.subscriptions, client.UUID)
					}
				}
			}
			hub.mu.Unlock()

		case <-hub.ctx.Done():
			return
		}
	}
}
