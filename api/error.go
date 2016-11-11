package api

import (
	"net/http"

	"github.com/anduintransaction/oauth-proxy/views"
	"gottb.io/goru"
	"gottb.io/gorux"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func Error(message string) *ErrorMessage {
	return &ErrorMessage{message}
}

var InternalServerError = Error("internal server error")

func RenderError(ctx *goru.Context, message string) {
	b, err := views.Error.Render(message)
	if err != nil {
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, InternalServerError)
	}
	goru.Ok(ctx, b)
}
