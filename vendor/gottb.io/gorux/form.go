package gorux

import "gottb.io/goru"

func Query(ctx *goru.Context, key string) string {
	return ctx.Request.URL.Query().Get(key)
}

func Form(ctx *goru.Context, key string) string {
	ctx.Request.ParseForm()
	return ctx.Request.Form.Get(key)
}
