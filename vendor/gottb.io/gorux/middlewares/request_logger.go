package middlewares

import (
	"gottb.io/goru"
	"gottb.io/goru/log"
)

type RequestLogger struct {
}

func (m *RequestLogger) Call(h goru.Handler) goru.Handler {
	return goru.HandlerFunc(func(ctx *goru.Context) {
		log.Debug("Request: ", ctx.Request.URL)
		h.Handle(ctx)
	})
}
