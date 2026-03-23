package pipeline

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// --- Pipeline статусы и шаги ---

type Step struct {
	Step int32  `json:"step"`
	Name string `json:"name"`
}

type Status string

const (
	Wait   Status = "wait"
	Done   Status = "done"
	Work   Status = "work"
	Cancel Status = "cancel"
	Error  Status = "error"
)

// PipelineMessage отправляется воркером в event.pipeline
type PipelineMessage struct {
	Status    Status `json:"status"`
	Step      int32  `json:"step"`
	WorkeUUID string `json:"worker_uuid"`
	UUID      string `json:"uuid"`
}

// --- Типы в очереди сообщений ---

type FileInMessage struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type MessageInQueue struct {
	ID        int32           `json:"id"`
	UUID      uuid.UUID       `json:"uuid"`
	Value     json.RawMessage `json:"value"`
	SendLater *time.Time      `json:"send_later,omitempty"`
	CreateAt  time.Time       `json:"create_at"`
	Files     []FileInMessage `json:"files"`
}

type SystemInQueue struct {
	Name string `json:"name"`
}

type PipelineInQueue struct {
	Step int32 `json:"step"`
}

// --- ConfigSchema для воркеров ---

type ConfigField struct {
	Type string `json:"type" validate:"required,alpha,oneof=numeric text ip"`
	Slug string `json:"slug" validate:"required,alpha"`
	Name string `json:"name" validate:"required,alphanum"`
}

// --- Регистрация воркера ---

type RegisterWorkerRequest struct {
	Token        string        `json:"token" validate:"required"`
	WorkerUUID   string        `json:"worker_uuid" validate:"required,uuid4"`
	Kind         string        `json:"kind" validate:"required"`
	NameKind     string        `json:"name_kind" validate:"required"`
	Type         string        `json:"type" validate:"required"`
	NameType     string        `json:"name_type" validate:"required"`
	ConfigSchema []ConfigField `json:"config_schema" validate:"required,omitnil"`
}

type RegisterWorkerResponse struct {
	Created bool                   `json:"created"`
	Config  map[string]interface{} `json:"config"`
	ID      int32                  `json:"id"`
}
