package telegram

type Schema struct {
	Message string `json:"message" validate:"required"`
}
