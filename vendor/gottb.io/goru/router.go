package goru

import (
	"fmt"
	"net/http"

	"gottb.io/goru/errors"

	"golang.org/x/net/context"
)

type Verb string

const (
	GET     Verb = "GET"
	POST    Verb = "POST"
	PUT     Verb = "PUT"
	DELETE  Verb = "DELETE"
	HEAD    Verb = "HEAD"
	OPTIONS Verb = "OPTIONS"
	PATCH   Verb = "PATCH"
	TRACE   Verb = "TRACE"
	CONNECT Verb = "CONNECT"
	ANY     Verb = "ANY"
)

type Router struct {
	tree            *routeTree
	notFoundHandler Handler
	panicHandler    Handler
}

func NewRouter() *Router {
	return &Router{
		tree:            newRouteTree(),
		notFoundHandler: NetHTTPHandlerFunc(http.NotFound),
		panicHandler:    HandlerFunc(defaultPanicFunc),
	}
}

func (r *Router) Register(verb Verb, route string, handler Handler) *Router {
	r.tree.add(verb, route, handler)
	return r
}

func (r *Router) Get(route string, handler Handler) *Router {
	return r.Register(GET, route, handler)
}

func (r *Router) Post(route string, handler Handler) *Router {
	return r.Register(POST, route, handler)
}

func (r *Router) Put(route string, handler Handler) *Router {
	return r.Register(PUT, route, handler)
}

func (r *Router) Delete(route string, handler Handler) *Router {
	return r.Register(DELETE, route, handler)
}

func (r *Router) Head(route string, handler Handler) *Router {
	return r.Register(HEAD, route, handler)
}

func (r *Router) Options(route string, handler Handler) *Router {
	return r.Register(OPTIONS, route, handler)
}

func (r *Router) Patch(route string, handler Handler) *Router {
	return r.Register(PATCH, route, handler)
}

func (r *Router) Trace(route string, handler Handler) *Router {
	return r.Register(TRACE, route, handler)
}

func (r *Router) Connect(route string, handler Handler) *Router {
	return r.Register(CONNECT, route, handler)
}

func (r *Router) Any(route string, handler Handler) *Router {
	return r.Register(ANY, route, handler)
}

func (r *Router) SetNotFoundHandler(h Handler) {
	r.notFoundHandler = h
}

func (r *Router) SetPanicHandler(h Handler) {
	r.panicHandler = h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	verb := Verb(req.Method)
	path := req.URL.Path
	h, args := r.tree.get(verb, path)
	netContext, cancel := context.WithCancel(context.Background())
	context := &Context{
		Request:        req,
		ResponseWriter: w,
		Params:         args,
		NetContext:     netContext,
	}
	defer func() {
		cancel()
		if rv := recover(); rv != nil {
			if rvErr, ok := rv.(*errors.Error); ok {
				context.Error = rvErr
			} else {
				context.Error = errors.NewError(fmt.Errorf("%v", rv), errors.StackTrace(4))
			}
			r.panicHandler.Handle(context)
		}
	}()
	if h == nil {
		r.notFoundHandler.Handle(context)
		return
	}
	h.Handle(context)
}

func defaultPanicFunc(c *Context) {
	http.Error(c.ResponseWriter, "500 Internal Server Error", http.StatusInternalServerError)
}
