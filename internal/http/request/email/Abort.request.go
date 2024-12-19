package email

type AbortRequest struct {
	UUID string `json:"uuid" validate:"required,uuid4"`
}
