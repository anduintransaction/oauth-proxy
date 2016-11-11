package api

import (
	"net/http"
	"net/url"

	"github.com/anduintransaction/oauth-proxy/proxy"
	"github.com/anduintransaction/oauth-proxy/service"
	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/log"
	"gottb.io/gorux"
)

func Begin(ctx *goru.Context) {
	p := proxy.GetProxy(ctx.Request.Host)
	if p == nil {
		gorux.ResponseJSON(ctx, http.StatusOK, Error("Anduin OAUTH proxy version "+service.Version()))
		return
	}
	user := service.CheckSession(ctx)
	if user != nil {
		goru.Redirect(ctx, "/")
		return
	}
	requestPath := gorux.Query(ctx, "request-path")
	if requestPath == "" {
		requestPath = "/"
	}
	requestURL, err := url.Parse(requestPath)
	if err != nil {
		log.Error(errors.Wrap(err))
		gorux.ResponseJSON(ctx, http.StatusBadRequest, Error("Invalid request URL"))
	}
	ctx.Request.URL = requestURL
	service.DoRedirect(ctx, p)
}
