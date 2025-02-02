package request

type GetMessageForRequest struct {
	ClientID     int `json:"client_id" validate:"required,min=1"`
	TypeWorkerID int `json:"type_worker_id" validate:"required,min=1"`
}
