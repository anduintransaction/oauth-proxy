package service

import (
	"encoding/base64"
	"encoding/json"

	"github.com/anduintransaction/oauth-proxy/provider"
	"github.com/anduintransaction/oauth-proxy/proxy"
	"gottb.io/goru"
	"gottb.io/goru/crypto"
	"gottb.io/goru/errors"
	"gottb.io/goru/log"
)

func DoRedirect(ctx *goru.Context, prox *proxy.Proxy) {
	prov := provider.GetProvider(prox.Provider)
	if prov == nil {
		log.Errorf("Proxy provider not found: %s", prox.Provider)
		goru.InternalServerError(ctx, []byte("InternalServerError"))
		return
	}
	randomState, err := generateRandomState()
	if err != nil {
		log.Error(err)
		goru.InternalServerError(ctx, []byte("InternalServerError"))
		return
	}
	proxy.AddState(randomState, prox, ctx.Request)
	redirectURI := prov.RedirectURI(prox, randomState)
	goru.Redirect(ctx, redirectURI)
}

func CheckWhitelist(ctx *goru.Context, prox *proxy.Proxy) bool {
	return prox.IsWhiteList(ctx.Request.Method, ctx.Request.URL.Path)
}

func CheckSession(ctx *goru.Context) *proxy.UserInfo {
	authCookie, err := ctx.Request.Cookie(proxy.Config.CookieName)
	if err != nil {
		log.Error(errors.Wrap(err))
		return nil
	}
	encrypted, err := base64.StdEncoding.DecodeString(authCookie.Value)
	if err != nil {
		log.Error(errors.Wrap(err))
		return nil
	}
	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		log.Error(err)
		return nil
	}
	session := &proxy.Session{}
	err = json.Unmarshal(decrypted, session)
	if err != nil {
		log.Error(errors.Wrap(err))
		return nil
	}
	log.Debugf("Got session for user %s", session.User)
	if proxy.Config.CheckVersion && session.Version != proxy.Config.Version {
		log.Debugf("Wrong version with user %s, expect %d but got %d", session.User, proxy.Config.Version, session.Version)
		return nil
	}
	return session.User
}

func ReverseProxy(ctx *goru.Context, prox *proxy.Proxy, user *proxy.UserInfo) {
	if user != nil {
		ctx.Request.Header.Add("X-Forwarded-User", user.Name)
		ctx.Request.Header.Add("X-Forwarded-Email", user.Email)
	}
	log.Debugf("Reverse proxy for %s to %s", prox.RequestHost, ctx.Request.URL.String())
	prox.ServeHTTP(ctx.ResponseWriter, ctx.Request)
}
