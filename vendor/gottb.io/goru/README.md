# Goru getting started guide

## Installation

```
cd $GOPATH/gottb.io
git clone git@github.com:gottb/goru.git
cd goru/tool/goru
go install
```

## Sample app: blog

### Create new app

```
cd $GOPATH/goru-sample
goru create app blog
```

### Run (currently empty) app

```
cd blog
goru run
```

By default, goru will create a HTTP server listened on port 8765, 
try accessing [http://localhost:8765](http://localhost:8765) now.

### What's inside an goru app?

* .gitignore <- git ignore file
* .watcher_ignore <- more on this later
* conf/app.toml <- main config file
* conf/private/app.toml <- private config file (will not be commited to git)
* generated <- ignore this
* generated.go <- also ignore this
* main.go <- begin with this file
* public <- public folder, put css/js/images... here

### Where's the rest (models, views, controllers)?

Goru is NOT a traditional web framework, you must decide the infrastructure yourself.

### The main.go file

```go
//go:generate goru generate asset public <- generate the "asset" in "public" folder.
//go:generate goru generate view <- generate necessary code for "views"
package main

import (
	"gottb.io/goru"
	"gottb.io/goru/crypto"
	"gottb.io/goru/handlers"
	"gottb.io/goru/log"
	"gottb.io/goru/session"
	"gottb.io/goru/packer"

	"gottb.io/goru-sample/blog/generated/assets/public"
)

func main() {
	r := goru.NewRouter() <- create new router
	r.Get("/assets/*file", &handlers.Assets{ <- register a route with a handler.
		Name: "file",
		Packs: map[string]*packer.Pack{
			"public": public.Get(),
		},
	})
	r.Get("/", goru.HandlerFunc(func(c *goru.Context) { <- register another route
		goru.Ok(c, []byte("Hello goru"))
	}))

	goru.StartWith(log.Start) <- start log service
	goru.StartWith(crypto.Start) <- start crypto service
	goru.StartWith(session.Start) <- start session service

	goru.StopWith(log.Stop) <- register stop hook for log
	goru.Run(r) <- run the app with this router.
}
```

### Router

A router maps a route (for example "/", "/home", "/login", "/user/edit/1") with a "Handler".

Create a new route with:

```go
r := goru.NewRouter()
```

A "Handler" is just an interface:

```go
type Handler interace {
    Handle(ctx *goru.Context) // More on Context later.
}
```

To register a static (fixed) route:

```go
r.Get("/hello", goru.HandlerFunc(func(ctx *goru.Context) {
    goru.Ok(ctx, []byte("Hello"))
}))
```

* `r.Get` will register a route with "GET" method.
* `goru.HandlerFunc` is an adapter that convert a `func(*goru.Context)` to a valid `goru.Handler`
* `goru.Ok(*goru.Context, []byte)` response with `200 OK` code and a `[]byte` as body.

The same route, but with different method:

```go
r.POST("/hello", goru.HandlerFunc(func(ctx *goru.Context) {
    goru.Ok(ctx, []byte("Post to Hello"))
}))
```

Try copy-paste these code to main.go and save.

### goru run command

After saving the file, goru will automatically detect a file has change and try to
rebuild the app. Run following command to see the changes in our app:

```
curl localhost:8765/hello
curl -X POST localhost:8765/hello
```

Something `goru run` has done behind the scene:

* Packing assets into go code so you don't have to copy your CSS/JS/Images file around
when deploying. A goru app has only a single binary file and a config file.
* Type checking view template files.
* Running in "watcher mode", which means changes in code files will trigger auto-rebuild.
However, it is so clever that changes in your asset files (CSS, JS...) will NOT trigger rebuild,
but still be updated in the running app. 

Some arguments can be added to `goru run` command:

* To run the HTTP Server on another port: `goru run 6666`
* To specify another config file: `goru run 6666 path/to/config.toml`

### Router (cont.)

Route with dynamic parameters:

```go
r.Get("/user/:name", goru.HandlerFunc(func(ctx *goru.Context){
    goru.Ok(ctx, []byte("Hello, " + ctx.Params["name"]))
}))
```

Route with dynamic parameters, with custom regular expression:

```go
r.Get("/post/$id<\\d+>", goru.HandlerFunc(func(ctx *goru.Context){
    goru.Ok(ctx, []byte("The id is: " + ctx.Params["id"]))
}))
```

* Access [http://localhost:8765/post/abc](http://localhost:8765/post/abc) to see a 404 Not Found response.

Route with dynamic parameters, spanning multiple parts:

```go
r.Get("/resources/*path", goru.HandlerFunc(func(ctx *goru.Context){
    goru.Ok(ctx, []byte(ctx.Params["path"]))
}))
```

* The response of [http://localhost:8765/resources/path/to/file.html](http://localhost:8765/resources/path/to/file.html) should be `path/to/file.html`

### Context

Something to pass around in a single request. Goru context is simply a struct:

```go
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         map[string]string
	Error          interface{}
	NetContext     context.Context // x/net/context.
}
```

### Middleware

Middleware solves everything: authentication, logging request, ...

Goru middleware is just a simple interface:

```go
type Middleware interface {
	Call(handler goru.Handler) goru.Handler
}
```

And an adapter:

```go
type MiddlewareFunc func(handler goru.Handler) goru.Handler
```

An example middleware that logs the response time for every request:

```go
func ResponseTimeMiddleware(handler goru.Handler) goru.Handler {
    return goru.HandlerFunc(func(ctx *goru.Context){
        begin := time.Now()
        handler.Handle(ctx)
        end := time.Now()
        log.Info("Response time: ", end.Sub(begin))
    })
}
```

### How to use middleware? Meet Action

To use middleware, goru provides a special Handler: `goru.Action`:

```go
type Action struct {
	Middlewares []goru.Middleware
	Handler     goru.Handler
}
```

An example action that use above middleware:

```go
r.Get("/logged", &goru.Action{
    Middlewares: []goru.Middleware{goru.MiddlewareFunc(ResponseTimeMiddleware)},
    Handler: goru.HandlerFunc(func(ctx *goru.Context){
        goru.Ok(ctx, []byte("Request logged"))
    }),
})
```
### View/Template

Goru makes go html/template **type safe**:

First we create a `views` folder to make thing easier, and define a template file inside,
let's name it `sample.tmpl.html`:

```
//import "gottb.io/goru"
//func(ctx *goru.Context, greeting string)
{{$greeting}}, {{index $ctx.Param "name"}}
```

Try save this file and you will see our app happily throw some errors:

```
views/sample.tmpl.html: type gottb.io/goru.Context has no field or method Param
{{index $ctx.Param "name"}}
```

As we can see, the almighty type inference feature of goru has detected a nasty error
in the template file (that maybe hard to be detected with human eyes). We shall fix this
silly error and save the file again:

```
{{$greeting}}, {{index $ctx.Params "name"}}
```

But how do we use this template file? We can look inside `views` folder and see another file
was created: `sample.tmpl.go`. A quick look of this file reveal something useful:

```go
var Sample = &sampleView{
    ... blah blah blah
}

func (v *sampleView) Render(ctx *goru.Context, greeting string) ([]byte, error) {
    blah blah ...
}
```

This means we can use Sample.Render and pass a `*goru.Context` and a `string` to that function
to make the template rendered.

Now register another route:

```go
r.Get("/greeting/:name", goru.HandlerFunc(func(ctx *goru.Context){
    b, _ := views.Sample.Render(ctx, "Hello")
    goru.Ok(ctx, b)
}))
```

Go to [http://localhost:8765/greeting/goru](http://localhost:8765/greeting/goru) to see expected result.

### Assets/Resources

By default, we can put assets/resources files inside `public` folder and those files can
be accessed with `assets/*file` route.

For example, created file `css/main.css` inside `public`, then the URL of this file is:
[http://localhost:8765/assets/public/css/main.css](http://localhost:8765/assets/public/css/main.css)

What's going on behind the scene? 
* First, goru create a `*goru.Packer` object that holds the path and content of every file
inside `public` folder, recursively (via the first go:generate in main.go).
* The `*goru.Packer` object now can be used to open and read file content, just likes `os.Open`.
* The `*goru.Packer` object can be found in `generated/assets/public/generated.go` file. This file is messy.
* The `Get` method in this file returns the much needed *goru.Packer object.
* goru shipped with a special Handler: `handlers.Asset`. This handler allows us to register a `*goru.Packer`
with a `string` key. In this example, we has registered the public packer with "public" key,
hence the URL "/assets/public/..."
* We can register another `*goru.Packer` with another key. This creates a powerful combo 
of flexibility and code reuse. For example, clone the repo [https://bitbucket.org/keimoonvie/assets](https://bitbucket.org/keimoonvie/assets)
and happily add `"jquery": jquery.Get()`.

### Configuration

Goru configuration file use TOML format. See [https://github.com/toml-lang/toml](https://github.com/toml-lang/toml) for syntax.

The main configuration file is `conf/app.toml`. The file `conf/private/app.toml` holds your
private configs, and will override existed configs in main file. The private config file 
is not committed to git.

As stated in `goru run` section, you can specify custom path to configuration file.

### goru dist

For deployment, run `goru dist` to create a single binary file and a config file inside `dist`
folder. Upload these 2 files to your server and we're done.  