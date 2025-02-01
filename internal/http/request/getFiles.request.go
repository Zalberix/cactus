package request

type GetFileRequest struct {
	Uuid string `json:"uuid" validate:"required,uuid"`
}
