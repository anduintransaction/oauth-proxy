package gorux

import (
	"html/template"
	"net/http"
	"os"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/log"
)

func InitErrorHandler(router *goru.Router) {
	router.SetNotFoundHandler(goru.HandlerFunc(NotFoundHandler))
	router.SetPanicHandler(goru.HandlerFunc(PanicHandler))
}

func Fail(ctx *goru.Context, err error) {
	var errorInfo *ErrorInfo
	var goruError *errors.Error
	var ok bool
	if goruError, ok = err.(*errors.Error); !ok {
		goruError = errors.NewError(err, errors.StackTrace(4))
	}
	if goruError.GetStack() != nil {
		log.Trace(goruError)
	}
	for _, h := range customErrorHandlerChain {
		errorInfo = h.Handle(ctx, goruError)
		if errorInfo != nil {
			break
		}
	}
	if errorInfo == nil {
		errorInfo = defaultErrorHandler(ctx, goruError)
	}
	if os.Getenv("RUNMODE") != "development" {
		errorInfo.Stack = nil
	}
	var body []byte
	var renderError error
	if errorInfo.Renderer != nil {
		body, renderError = errorInfo.Renderer.Render(ctx, errorInfo)
	} else {
		body, renderError = defaultErrorRenderer(ctx, errorInfo)
	}
	if renderError != nil {
		// WTF?
		log.Trace(renderError)
		goru.InternalServerError(ctx, []byte("Internal Server Error"))
	} else {
		goru.Response(ctx, errorInfo.StatusCode, body)
	}
}

type ErrorInfo struct {
	StatusCode   int
	Status       string
	ErrorCode    int
	ErrorMessage string
	Stack        []string
	Renderer     Renderer
}

type ErrorHandler interface {
	Handle(ctx *goru.Context, err *errors.Error) *ErrorInfo
}

func AddErrorHandler(h ErrorHandler) {
	customErrorHandlerChain = append(customErrorHandlerChain, h)
}

func AddErrorHandlers(h []ErrorHandler) {
	customErrorHandlerChain = append(customErrorHandlerChain, h...)
}

func AddErrorTemplate(statusCode int, renderer Renderer) {
	errorTemplates[statusCode] = renderer
}

type errorHandlerChain []ErrorHandler

var customErrorHandlerChain = errorHandlerChain{}

type Renderer interface {
	Render(ctx *goru.Context, err *ErrorInfo) ([]byte, error)
}

func NotFound(ctx *goru.Context, msg string) {
	Fail(ctx, &httpError{http.StatusNotFound, msg})
}

func Forbidden(ctx *goru.Context, msg string) {
	Fail(ctx, &httpError{http.StatusForbidden, msg})
}

func BadRequest(ctx *goru.Context, msg string) {
	Fail(ctx, &httpError{http.StatusBadRequest, msg})
}

func NotFoundHandler(ctx *goru.Context) {
	err := &httpError{http.StatusNotFound, "404 Not Found"}
	// Should ignore stack trace here
	Fail(ctx, errors.NewError(err, nil))
}

func PanicHandler(ctx *goru.Context) {
	Fail(ctx, ctx.Error)
}

func defaultErrorHandler(ctx *goru.Context, err *errors.Error) *ErrorInfo {
	errorInfo := &ErrorInfo{
		ErrorMessage: err.Error(),
		Stack:        err.GetStack(),
	}
	switch underlying := err.Underlying().(type) {
	case *httpError:
		errorInfo.StatusCode = underlying.statusCode
	default:
		errorInfo.StatusCode = http.StatusInternalServerError
	}
	errorInfo.Status = http.StatusText(errorInfo.StatusCode)
	errorInfo.ErrorCode = errorInfo.StatusCode
	return errorInfo
}

var defaultErrorTemplate = template.Must(template.New("gorux/default-error-template").Parse(`
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<title>{{.Status}}</title>
	</head>
	<body>
		<h1>{{.ErrorMessage}}</h1>
		{{if .Stack}}
			<h3>Stack trace</h3>
			{{range .Stack}}
				<p>{{.}}</p>
			{{end}}
		{{end}}		
	</body>
</html>
`))

var errorTemplates = map[int]Renderer{}

func defaultErrorRenderer(ctx *goru.Context, errorInfo *ErrorInfo) ([]byte, error) {
	if template, ok := errorTemplates[errorInfo.StatusCode]; ok {
		return template.Render(ctx, errorInfo)
	}
	return ExecuteTemplate(defaultErrorTemplate, errorInfo)
}

type httpError struct {
	statusCode int
	msg        string
}

func (e *httpError) Error() string {
	return e.msg
}
