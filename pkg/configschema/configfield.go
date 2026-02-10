package configschema

type ConfigField struct {
	// TODO oneof реализовать на стороне API, тут это не благодарное дело
	Type string `json:"type" validate:"required,alpha,oneof=numeric text ip"`
	Slug string `json:"slug" validate:"required,alpha"`
	Name string `json:"name" validate:"required,alphanum"`
	// Validation string `json:"validation" validate:"required"` // TODO добавить правило для валидации правил валидатора.
}
