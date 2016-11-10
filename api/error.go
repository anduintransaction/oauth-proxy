package api

type ErrorMessage struct {
	Message string `json:"message"`
}

func Error(message string) *ErrorMessage {
	return &ErrorMessage{message}
}

var InternalServerError = Error("internal server error")
