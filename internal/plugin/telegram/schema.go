package telegram

type Schema struct {
	Title    string   `json:"title"`    // Заголовок
	Message  string   `json:"message"`  // Cообщение
	Subjects []string `json:"subjects"` // Кому
}
