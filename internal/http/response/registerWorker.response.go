package response

type RegisterWorkerResponse struct {
	Created bool                   `json:"created"`
	Config  map[string]interface{} `json:"config"`
	ID      int32                  `json:"id"`
}
