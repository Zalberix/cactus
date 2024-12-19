package message

type GetMessageByRequest struct {
	ClientId  int `json:"client_id" validate:"required,min=1"`
	ProcessId int `json:"process_id" validate:"required,min=1"`
}
