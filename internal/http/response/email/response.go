package email

type ResponceStatus bool

type ResponceOk struct {
	Success ResponceStatus `json:"success"`
	Data    any            `json:"data"`
}

type ResponceError struct {
	Success ResponceStatus    `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func CreateResponseOK(data any) ResponceOk {
	return ResponceOk{
		Success: true,
		Data:    data,
	}
}

func CreateResponseError(message string, errors map[string]string) ResponceError {
	return ResponceError{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}
