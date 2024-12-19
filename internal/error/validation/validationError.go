package validationerror

import (
	"strings"

	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ru_translations "github.com/go-playground/validator/v10/translations/ru"
)

// TODO оптимизировать, каждый раз создается переводчик для валидатора
func ValidateStructure(structure any) map[string]string {
	ru := ru.New()
	uni := ut.New(ru)

	trans, _ := uni.GetTranslator("ru")

	validate := validator.New()
	ru_translations.RegisterDefaultTranslations(validate, trans)

	err := validate.Struct(structure)
	if err != nil {

		errs := err.(validator.ValidationErrors)

		validationErrors := make(map[string]string, len(errs))

		for _, e := range errs {
			validationErrors[genNameFromNamespaceError(e.StructNamespace())] = e.Translate(trans)
		}
		return validationErrors
	}
	return nil

}

func genNameFromNamespaceError(namespace string) string {
	var newNamespace []string
	for i, part := range strings.Split(namespace, ".") {
		if i == 0 {
			continue
		}
		newNamespace = append(newNamespace, strings.ToLower(part))
	}
	return strings.Join(newNamespace, ".")
}
