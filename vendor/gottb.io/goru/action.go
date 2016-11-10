package goru

type Action struct {
	Middlewares []Middleware
	Handler     Handler
}

func (action *Action) Handle(c *Context) {
	h := action.Handler
	for i := len(action.Middlewares) - 1; i >= 0; i-- {
		h = action.Middlewares[i].Call(h)
	}
	h.Handle(c)
}
