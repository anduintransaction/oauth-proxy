//go:generate goru generate view
package main

import (
	"github.com/anduintransaction/oauth-proxy/api"
	"github.com/anduintransaction/oauth-proxy/proxy"
	"gottb.io/goru"
	"gottb.io/goru/crypto"
	"gottb.io/goru/log"
	"gottb.io/goru/session"
	_ "gottb.io/gorux/functions"
)

func main() {
	r := goru.NewRouter()
	r.Any("/**", goru.HandlerFunc(api.Main))
	r.Get("/oauth2/callback", goru.HandlerFunc(api.Callback))
	r.Get("/oauth2/login", goru.HandlerFunc(api.Login))
	r.Get("/oauth2/begin", goru.HandlerFunc(api.Begin))
	r.Get("/favicon.ico", goru.HandlerFunc(api.Favicon))

	goru.StartWith(log.Start)
	goru.StartWith(crypto.Start)
	goru.StartWith(session.Start)
	goru.StartWith(proxy.Start)

	goru.StopWith(log.Stop)
	goru.StopWith(proxy.Stop)
	goru.Run(r)
}
