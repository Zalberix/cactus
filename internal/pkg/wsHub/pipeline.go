package wshub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	dto "cactus/internal/DTO"
	"cactus/internal/pkg/pipeline"
)

type PipelineMessage struct {
	Status    pipeline.Status `json:"status"`
	Step      int32           `json:"step"`
	WorkeUUID string          `json:"worker_uuid"`
	UUID      string          `json:"uuid"`
}

type PipelineClient struct {
	Hub     *PipelineHub
	Conn    *websocket.Conn
	UUID    string
	message chan []byte

	ctx context.Context
}

func (c *PipelineClient) Listen() {
	for message := range c.message {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
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
	// CancelPipeline(ctx context.Context, uuid string, step pipeline.Step)
	// ErrorPipeline(ctx context.Context, uuid string, step pipeline.Step)
}

type PipelineHub struct {
	clients       map[*PipelineClient]bool
	subscriptions map[string]map[*PipelineClient]bool // подписки по UUID

	upgrader *websocket.Upgrader
	rdb      *redis.Client
	ctx      context.Context

	Service    ServicePipelineHub
	Register   chan *PipelineClient
	Unregister chan *PipelineClient

	mu sync.Mutex
}

func NewPipelineHub(ctx context.Context, rdb *redis.Client, service ServicePipelineHub) *PipelineHub {
	return &PipelineHub{
		clients:       make(map[*PipelineClient]bool),
		subscriptions: make(map[string]map[*PipelineClient]bool),

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		rdb:     rdb,
		ctx:     ctx,
		mu:      sync.Mutex{},
		Service: service,

		Register:   make(chan *PipelineClient),
		Unregister: make(chan *PipelineClient),
	}
}

func (hub *PipelineHub) NewClient(ctx context.Context, uuid string, conn *websocket.Conn) *PipelineClient {
	client := &PipelineClient{
		ctx:     ctx,
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

func (hub *PipelineHub) Send(pm PipelineMessage) {
	hub.mu.Lock()

	newPipeline, err := hub.Service.UpdateStatusPipeline(
		hub.ctx,
		pm.UUID,
		pm.Step,
		pm.Status,
		pm.WorkeUUID,
	)

	slog.Error(
		"Сообщение перед отправкой всем",
		slog.Any("message", pm),
	)

	if err != nil {
		slog.Error(
			"Переход к шагу не возможен",
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
	hub.mu.Unlock()
}

// Run запускает обработчик событий
func (hub *PipelineHub) Run() {
	go hub.ReadStream(hub.ctx)

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
		}
	}
}

type Message struct {
	UUID    string
	Message PipelineMessage
}

func (hub *PipelineHub) ReadStream(ctx context.Context) error {
	streamName := "event:pipeline"
	groupName := "pipeline"
	consumerName := "cactus"

	err := hub.createStreamIfNotExist(ctx, streamName, groupName)
	if err != nil {
		return fmt.Errorf("ошибка создания стрима %v: %w", streamName, err)
	}

	_, err = hub.rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("Ошибка создания группы в Redis Stream: %v", err)
	}

	for {
		msgs, err := hub.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    60,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				log.Printf("Ошибка чтения из стрима: %v", err)
			}
			continue
		}

		for _, stream := range msgs {
			for _, m := range stream.Messages {
				messageJSON, ok := m.Values["message"].(string)

				if !ok {
					slog.Error(
						"Сообщение от воркера отсутсвует",
						slog.Any("messageJSON", messageJSON))
					continue
				}

				var pipelineMessage PipelineMessage

				if err = json.Unmarshal([]byte(messageJSON), &pipelineMessage); err != nil {
					slog.Error(
						"Сообщение от воркера невозможно обработать",
						slog.Any("messageJSON", messageJSON),
						slog.Any("err", err.Error()),
					)
					continue
				}

				hub.Send(pipelineMessage)

				// Подтверждение обработки
				hub.rdb.XAck(ctx, streamName, groupName, m.ID)
			}
		}
	}
}

func (hub *PipelineHub) createStreamIfNotExist(ctx context.Context, key string, group string) error {
	typeRes, err := hub.rdb.Type(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("ошибка получения типа ключа %s: %w", key, err)
	}

	switch typeRes {
	case "none":
		if group != "" {
			err = hub.rdb.XGroupCreateMkStream(ctx, key, group, "$").Err()
			if err != nil {
				return fmt.Errorf("ошибка создания группы %s в стриме %s: %w", group, key, err)
			}
		} else {
			_, err = hub.rdb.XAdd(ctx, &redis.XAddArgs{
				Stream: key,
				Values: map[string]interface{}{"init": "stream"},
			}).Result()
			if err != nil {
				return fmt.Errorf("ошибка создания стрима %s: %w", key, err)
			}
		}
	case "stream":
		info, err := hub.rdb.XInfoStream(ctx, key).Result()
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
		err = hub.rdb.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("ошибка удаления ключа %s: %w", key, err)
		}
		return hub.createStreamIfNotExist(ctx, key, group)
	}

	return nil
}
