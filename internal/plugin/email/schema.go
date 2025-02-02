package email

// TODO ДОБАВИТЬ валидационные теги!
type Schema struct {
	Title   string `json:"title" validate:"required,max=255"` // Заголовок
	Message string `json:"message" validate:"required"`       // Cообщение
	Subject string `json:"subject" validate:"required"`       // Кому

	// Тип генерации сообщения по шаблона или как текст
	TypeMessage *string           `json:"type_message"`
	Data        map[string]string `json:"data"` // Данные для постановки
}
