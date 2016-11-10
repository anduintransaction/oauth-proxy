package goru

import "net/http"

func Response(ctx *Context, status int, data []byte) {
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write(data)
}

func Ok(ctx *Context, data []byte) {
	Response(ctx, http.StatusOK, data)
}

func BadRequest(ctx *Context, data []byte) {
	Response(ctx, http.StatusBadRequest, data)
}

func Unauthorized(ctx *Context, data []byte) {
	Response(ctx, http.StatusUnauthorized, data)
}

func PaymentRequired(ctx *Context, data []byte) {
	Response(ctx, http.StatusPaymentRequired, data)
}

func Forbidden(ctx *Context, data []byte) {
	Response(ctx, http.StatusForbidden, data)
}

func NotFound(ctx *Context, data []byte) {
	Response(ctx, http.StatusNotFound, data)
}

func MethodNotAllowed(ctx *Context, data []byte) {
	Response(ctx, http.StatusMethodNotAllowed, data)
}

func InternalServerError(ctx *Context, data []byte) {
	Response(ctx, http.StatusInternalServerError, data)
}

func NotImplemented(ctx *Context, data []byte) {
	Response(ctx, http.StatusNotImplemented, data)
}

func BadGateway(ctx *Context, data []byte) {
	Response(ctx, http.StatusBadGateway, data)
}

func Redirect(ctx *Context, url string) {
	http.Redirect(ctx.ResponseWriter, ctx.Request, url, http.StatusFound)
}

func SetCookie(ctx *Context, cookie *http.Cookie) {
	http.SetCookie(ctx.ResponseWriter, cookie)
}
