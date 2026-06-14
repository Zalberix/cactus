package configpub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/zalberix/cactus/apps/manager/internal/natssubjects"
	"github.com/zalberix/cactus/libs/storage/db"
)

type Payload struct {
	OrganizationID int32           `json:"organization_id"`
	WorkTypeID     int32           `json:"work_type_id"`
	SchemaID       int32           `json:"schema_id"`
	RevisionID     int32           `json:"revision_id"`
	ConfigHash     string          `json:"config_hash"`
	SettingsData   json.RawMessage `json:"settings_data"`
}

type Bus interface {
	PublishJS(ctx context.Context, subject string, data any) error
}

type requestBus interface {
	Bus
	Publish(subject string, data any) error
	SubscribeMsg(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error)
}

type Store interface {
	GetWorkerSettingsRevisionByID(ctx context.Context, id int32) (db.WorkerSettingsRevision, error)
	GetWorkerSettingsSchemaByID(ctx context.Context, id int32) (db.WorkerSettingsSchema, error)
}

type Service struct {
	store Store
	bus   Bus
	sub   *nats.Subscription
}

func New(store Store, bus Bus) *Service {
	return &Service{store: store, bus: bus}
}

func (s *Service) PublishRevision(ctx context.Context, orgID, workTypeID, schemaID, revisionID int32, settings []byte) error {
	payload := Payload{
		OrganizationID: orgID,
		WorkTypeID:     workTypeID,
		SchemaID:       schemaID,
		RevisionID:     revisionID,
		ConfigHash:     hash(settings),
		SettingsData:   json.RawMessage(settings),
	}
	return s.bus.PublishJS(ctx, natssubjects.Config(orgID, workTypeID, revisionID), payload)
}

func (s *Service) Start(ctx context.Context) error {
	bus, ok := s.bus.(requestBus)
	if !ok {
		return fmt.Errorf("config publisher bus does not support request subscriptions")
	}
	sub, err := bus.SubscribeMsg("config.request.>", func(msg *nats.Msg) {
		s.handleRequest(ctx, bus, msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe config requests: %w", err)
	}
	s.sub = sub
	return nil
}

func (s *Service) Stop(context.Context) error {
	if s.sub == nil {
		return nil
	}
	return s.sub.Unsubscribe()
}

func (s *Service) handleRequest(ctx context.Context, bus requestBus, msg *nats.Msg) {
	orgID, workTypeID, revisionID, err := parseConfigRequestSubject(msg.Subject)
	if err != nil {
		return
	}
	revision, err := s.store.GetWorkerSettingsRevisionByID(ctx, revisionID)
	if err != nil {
		return
	}
	schema, err := s.store.GetWorkerSettingsSchemaByID(ctx, revision.WorkerSettingsSchemaID)
	if err != nil {
		return
	}
	if schema.WorkTypeID != workTypeID {
		return
	}
	payload := Payload{
		OrganizationID: orgID,
		WorkTypeID:     workTypeID,
		SchemaID:       schema.ID,
		RevisionID:     revision.ID,
		ConfigHash:     hash(revision.SettingsData),
		SettingsData:   json.RawMessage(revision.SettingsData),
	}
	_ = bus.PublishJS(ctx, natssubjects.Config(orgID, workTypeID, revisionID), payload)
	if msg.Reply != "" {
		_ = bus.Publish(msg.Reply, payload)
	}
}

func parseConfigRequestSubject(subject string) (int32, int32, int32, error) {
	parts := strings.Split(subject, ".")
	if len(parts) != 8 ||
		parts[0] != "config" ||
		parts[1] != "request" ||
		parts[2] != "org" ||
		parts[4] != "work_type" ||
		parts[6] != "revision" {
		return 0, 0, 0, fmt.Errorf("invalid config request subject: %s", subject)
	}
	orgID, err := parseInt32(parts[3])
	if err != nil {
		return 0, 0, 0, err
	}
	workTypeID, err := parseInt32(parts[5])
	if err != nil {
		return 0, 0, 0, err
	}
	revisionID, err := parseInt32(parts[7])
	if err != nil {
		return 0, 0, 0, err
	}
	return orgID, workTypeID, revisionID, nil
}

func parseInt32(raw string) (int32, error) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse id %q: %w", raw, err)
	}
	return int32(id), nil
}

func hash(settings []byte) string {
	sum := sha256.Sum256(settings)
	return "sha256:" + hex.EncodeToString(sum[:])
}
