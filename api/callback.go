package api

import (
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
		RenderError(ctx, "State is required")
		return
	}
	state := proxy.GetState(stateName)
	if state == nil {
		RenderError(ctx, "State not found or expired")
		return
	}
	prov := provider.GetProvider(state.Proxy.Provider)
	if prov == nil {
		log.Errorf("Provider not found: %s", state.Proxy.Provider)
		RenderError(ctx, InternalServerError.Message)
		return
	}
	code := gorux.Query(ctx, "code")
	if code == "" {
		log.Errorf("Error found in callback: %s", ctx.Request.URL.String())
		RenderError(ctx, prov.ErrorString(ctx.Request))
		return
	}
	token, err := prov.RequestToken(state, code)
	if err != nil {
		log.Error(err)
		RenderError(ctx, "Cannot request token")
		return
	}
	user, err := prov.VerifyUser(state, token)
	if err != nil {
		log.Error(err)
		RenderError(ctx, "Cannot verify user")
		return
	}
	if user == nil {
		RenderError(ctx, "Unauthorized user")
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
