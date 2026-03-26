package message

// SendMessageRequest — запрос на отправку сообщения (per D-10).
// JSON body: {"workflow_id": 1, "value": {...}, "external_id": "ext-123"}
type SendMessageRequest struct {
	WorkflowID int32          `json:"workflow_id" binding:"required"`
	Value      map[string]any `json:"value" binding:"required"`
	ExternalID string         `json:"external_id"`
}

// SendMessageResponse — ответ на успешную отправку (per D-14).
type SendMessageResponse struct {
	MessageID     int32  `json:"message_id"`
	WorkflowRunID int32  `json:"workflow_run_id"`
	Status        string `json:"status"` // "running"
}
