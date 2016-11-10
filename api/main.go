package api

import (
	"net/http"

	"github.com/anduintransaction/oauth-proxy/proxy"
	"github.com/anduintransaction/oauth-proxy/service"
	"gottb.io/goru"
	"gottb.io/gorux"
)

func Main(ctx *goru.Context) {
	p := proxy.GetProxy(ctx.Request.Host)
	if p == nil {
		gorux.ResponseJSON(ctx, http.StatusOK, Error("Anduin OAUTH proxy version "+service.Version()))
		return
	}
	user := service.CheckSession(ctx)
	if user != nil {
		service.ReverseProxy(ctx, p, user)
		return
	}
	service.DoRedirect(ctx, p)
}

func Favicon(ctx *goru.Context) {
	p := proxy.GetProxy(ctx.Request.Host)
	if p == nil {
		gorux.ResponseJSON(ctx, http.StatusNotFound, Error("not found"))
		return
	}
	user := service.CheckSession(ctx)
	if user != nil {
		service.ReverseProxy(ctx, p, user)
		return
	}
	gorux.ResponseJSON(ctx, http.StatusNotFound, Error("not found"))
}
