package api

import (
	"net/http"
	"net/url"

	"github.com/anduintransaction/oauth-proxy/provider"
	"github.com/anduintransaction/oauth-proxy/proxy"

	"gottb.io/goru"
	"gottb.io/goru/log"
	"gottb.io/gorux"
)

func Callback(ctx *goru.Context) {
	stateName := gorux.Query(ctx, "state")
	if stateName == "" {
		gorux.ResponseJSON(ctx, http.StatusBadRequest, Error("state is required"))
		return
	}
	state := proxy.GetState(stateName)
	if state == nil {
		gorux.ResponseJSON(ctx, http.StatusUnauthorized, Error("state not found or expired"))
		return
	}
	prov := provider.GetProvider(state.Proxy.Provider)
	if prov == nil {
		log.Errorf("Provider not found: %s", state.Proxy.Provider)
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, InternalServerError)
		return
	}
	code := gorux.Query(ctx, "code")
	if code == "" {
		log.Errorf("Error found in callback: %s", ctx.Request.URL.String())
		gorux.ResponseJSON(ctx, http.StatusBadRequest, Error(prov.ErrorString(ctx.Request)))
		return
	}
	token, err := prov.RequestToken(state, code)
	if err != nil {
		log.Error(err)
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, Error("cannot request token"))
		return
	}
	user, err := prov.VerifyUser(state, token)
	if err != nil {
		log.Error(err)
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, Error("cannot verify user"))
		return
	}
	if user == nil {
		gorux.ResponseJSON(ctx, http.StatusUnauthorized, Error("unauthorized user"))
		return
	}

	state.User = user
	redirectURL := url.URL{
		Scheme: state.Proxy.Scheme,
		Host:   state.Proxy.RequestHost,
		Path:   "/oauth2/login",
	}
	values := make(url.Values)
	values.Set("state", stateName)
	redirectURL.RawQuery = values.Encode()
	goru.Redirect(ctx, redirectURL.String())
}
