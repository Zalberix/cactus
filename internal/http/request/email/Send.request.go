package email

import "time"

type SendRequest struct {
	Title        string             `schema:"title" validate:"required,max=255"`
	Message      string             `schema:"message" validate:"required"`
	PrioritySlug string             `schema:"priority" validate:"required,min=0,max=5"`
	Subjects     []string           `schema:"subjects[]" validate:"required,min=1"`
	TypeMessage  *string            `schema:"type_message" validate:"omitnil,oneof=text html template"`
	Data         *map[string]string `schema:"data[]" validate:"required_if=Type template,omitempty,min=1"`
	SendLater    *time.Time         `schema:"send_later" validate:"omitnil,required"` // Сомнительно но окейd
	Theme        *string            `schema:"theme" validate:"required"`
	Headers      *map[string]string `schema:"headers[]" validate:"omitnil"`
	Files        []*struct {
		Title *string `schema:"title" validate:"required,min=5,max=255"`
		Name  *string `schema:"name" validate:"required"`
		Type  *string `schema:"type" validate:"required,oneof=bin base64"`
	} `schema:"files[]" validate:"omitempty,dive"`
}
