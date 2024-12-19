package email

type SendResponse struct {
	Status string `json:"status"`
	UUID   string `json:"uuid"`
}
