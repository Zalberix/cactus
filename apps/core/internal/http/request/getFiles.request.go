package request

type GetFileRequest struct {
	UUID string `json:"uuid" validate:"required,uuid"`
}
