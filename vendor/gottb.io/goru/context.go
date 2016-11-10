package goru

import (
	"net/http"

	"gottb.io/goru/errors"

	"golang.org/x/net/context"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         map[string]string
	Error          *errors.Error
	NetContext     context.Context
}
