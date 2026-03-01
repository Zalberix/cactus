package request

type GetStatusMessageRequest struct {
	UUID string `json:"uuid" validate:"required,uuid4"`
}
