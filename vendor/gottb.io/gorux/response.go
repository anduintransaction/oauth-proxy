package gorux

import (
	"encoding/json"

	"gottb.io/goru"
)

func AddHeader(ctx *goru.Context, key, value string) {
	ctx.ResponseWriter.Header().Add(key, value)
}

func ResponseJSON(ctx *goru.Context, statusCode int, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	AddHeader(ctx, "Content-Type", "application/json")
	goru.Response(ctx, statusCode, b)
	return nil
}

func ResponseJSONPretty(ctx *goru.Context, statusCode int, data interface{}) error {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	AddHeader(ctx, "Content-Type", "application/json")
	goru.Response(ctx, statusCode, b)
	return nil
}
