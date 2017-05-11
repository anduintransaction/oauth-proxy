package api

import (
	"net/http"

	"github.com/anduintransaction/oauth-proxy/proxy"
	"github.com/anduintransaction/oauth-proxy/service"
	"github.com/anduintransaction/oauth-proxy/views"
	"gottb.io/goru"
	"gottb.io/goru/log"
	"gottb.io/gorux"
)

func Main(ctx *goru.Context) {
	p := proxy.GetProxy(ctx.Request.Host)
	if p == nil {
		gorux.ResponseJSON(ctx, http.StatusOK, Error("Anduin OAUTH proxy version "+service.Version()))
		return
	}
	if service.CheckWhitelist(ctx, p) {
		service.ReverseProxy(ctx, p, nil)
		return
	}
	user := service.CheckSession(ctx)
	if user != nil {
		service.ReverseProxy(ctx, p, user)
		return
	}
	content, err := views.Index.Render(ctx.Request.URL.String())
	if err != nil {
		log.Error(err)
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, InternalServerError)
	}
	goru.Unauthorized(ctx, content)
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
