package request

import configschema "cactus/internal/pkg/configSchema"

type RegisterWorkerRequest struct {
	Token        string                     `json:"token" validate:"required"`
	WorkerUUID   string                     `json:"worker_uuid" validate:"required,uuid4"`
	Kind         string                     `json:"kind" validate:"required,alpha"`
	Type         string                     `json:"type" validate:"required,alpha"`
	ConfigSchema []configschema.ConfigField `json:"config_schema" validate:"required,omitnil"`
}
