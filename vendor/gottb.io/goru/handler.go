package goru

import "net/http"

type Handler interface {
	Handle(c *Context)
}

type HandlerFunc func(c *Context)

func (hf HandlerFunc) Handle(c *Context) {
	hf(c)
}

type NetHTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

func (h NetHTTPHandlerFunc) Handle(c *Context) {
	h(c.ResponseWriter, c.Request)
}
