package goru

type Middleware interface {
	Call(handler Handler) Handler
}

type MiddlewareFunc func(handler Handler) Handler

func (m MiddlewareFunc) Call(handler Handler) Handler {
	return m(handler)
}
