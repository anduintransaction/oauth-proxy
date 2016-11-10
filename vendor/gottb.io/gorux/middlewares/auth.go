package middlewares

import "gottb.io/goru"

type Auth struct {
	Authenticate    func(ctx *goru.Context) bool
	Unauthenticated func(ctx *goru.Context)
}

func (m *Auth) Call(handler goru.Handler) goru.Handler {
	return goru.HandlerFunc(func(ctx *goru.Context) {
		if m.Authenticate(ctx) {
			handler.Handle(ctx)
		} else {
			m.Unauthenticated(ctx)
		}
	})
}
