package contracts

import configschema "cactus/pkg/configschema"

type RegisterWorkerRequest struct {
	Token        string                     `json:"token" validate:"required"`
	WorkerUUID   string                     `json:"worker_uuid" validate:"required,uuid4"`
	Kind         string                     `json:"kind" validate:"required"`
	NameKind     string                     `json:"name_kind" validate:"required"`
	Type         string                     `json:"type" validate:"required"`
	NameType     string                     `json:"name_type" validate:"required"`
	ConfigSchema []configschema.ConfigField `json:"config_schema" validate:"required,omitnil"`
}

type RegisterWorkerResponse struct {
	Created bool                   `json:"created"`
	Config  map[string]interface{} `json:"config"`
	ID      int32                  `json:"id"`
}
