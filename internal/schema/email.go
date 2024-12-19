package schema

type Email struct {
	TypeMessage string            `json:"type_message"` // Тип генерации сообщения по шаблона или как текст
	Title       string            `json:"title"`        // Заголовок
	Message     string            `json:"message"`      // Cообщение
	Subjects    []string          `json:"subjects"`     // Кому
	Data        map[string]string `json:"data"`         // Данные для постановки
	Theme       string            `json:"theme"`        // Тема письма
}
