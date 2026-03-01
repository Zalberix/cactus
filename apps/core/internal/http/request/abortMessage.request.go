package request

type AbortMessageRequest struct {
	UUID string `json:"uuid" validate:"required,uuid4"`
}
