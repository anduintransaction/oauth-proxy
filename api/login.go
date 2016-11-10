package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/anduintransaction/oauth-proxy/proxy"

	"gottb.io/goru"
	"gottb.io/goru/crypto"
	"gottb.io/goru/errors"
	"gottb.io/goru/log"
	"gottb.io/gorux"
)

func Login(ctx *goru.Context) {
	stateName := gorux.Query(ctx, "state")
	if stateName == "" {
		gorux.ResponseJSON(ctx, http.StatusBadRequest, Error("state is required"))
		return
	}
	state := proxy.AcquireState(stateName)
	if state == nil {
		gorux.ResponseJSON(ctx, http.StatusUnauthorized, Error("state not found or expired"))
		return
	}
	if state.User == nil {
		gorux.ResponseJSON(ctx, http.StatusUnauthorized, Error("user was not authenticated"))
		return
	}
	session := &proxy.Session{
		User:    state.User,
		Version: proxy.Config.Version,
	}
	cookieContent, err := json.Marshal(session)
	if err != nil {
		log.Error(errors.Wrap(err))
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, InternalServerError)
		return
	}
	encryptedContent, err := crypto.Encrypt(cookieContent)
	if err != nil {
		log.Error(err)
		gorux.ResponseJSON(ctx, http.StatusInternalServerError, InternalServerError)
		return
	}
	goru.SetCookie(ctx, &http.Cookie{
		Domain:  state.Proxy.RequestHost,
		Name:    proxy.Config.CookieName,
		Value:   base64.StdEncoding.EncodeToString(encryptedContent),
		Path:    "/",
		Expires: time.Now().Add(time.Duration(proxy.Config.CookieTimeout) * time.Second),
	})
	goru.Redirect(ctx, state.Request.URL.String())
}
