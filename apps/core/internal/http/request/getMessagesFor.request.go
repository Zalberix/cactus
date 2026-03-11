package request

type GetMessageForRequest struct {
	ClientID     int32 `json:"client_id" validate:"required,min=1"`
	TypeWorkerID int32 `json:"type_worker_id" validate:"required,min=1"`
}
