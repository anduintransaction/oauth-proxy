package funcs

import (
	"gottb.io/goru"
	"gottb.io/goru/view"
)

func flash(ctx *goru.Context, key string) string {
	return goru.GetSession(ctx).GetFlash(key)
}

func init() {
	view.Funcs(map[string]interface{}{
		"flash": flash,
	})
}
