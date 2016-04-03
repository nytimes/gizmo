package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"

	"github.com/NYTimes/gizmo/config"
)

// Router is an interface to wrap different router types to be embedded within
// Gizmo server.Server implementations.
type Router interface {
	Handle(string, string, http.Handler)
	HandleFunc(string, string, func(http.ResponseWriter, *http.Request))
	ServeHTTP(http.ResponseWriter, *http.Request)
	SetNotFoundHandler(http.Handler)
}

// NewRouter will return the router specified by the server
// config. If no Router value is supplied, the server
// will default to using Gorilla mux.
func NewRouter(cfg *config.Server) Router {
	switch cfg.RouterType {
	case "gorilla":
		return &GorillaRouter{mux.NewRouter()}
	case "httprouter", "fast":
		return &FastRouter{httprouter.New()}
	default:
		return &GorillaRouter{mux.NewRouter()}
	}
}

// GorillaRouter is a Router implementation for the Gorilla web toolkit's `mux.Router`.
type GorillaRouter struct {
	mux *mux.Router
}

// Handle will call the Gorilla web toolkit's Handle().Method() methods.
func (g *GorillaRouter) Handle(method, path string, h http.Handler) {
	g.mux.Handle(path, h).Methods(method)
}

// HandleFunc will call the Gorilla web toolkit's HandleFunc().Method() methods.
func (g *GorillaRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.mux.HandleFunc(path, h).Methods(method)
}

// SetNotFoundHandler will set the Gorilla mux.Router.NotFoundHandler.
func (g *GorillaRouter) SetNotFoundHandler(h http.Handler) {
	g.mux.NotFoundHandler = h
}

// ServeHTTP will call Gorilla mux.Router.ServerHTTP directly.
func (g *GorillaRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// FastRouter is a Router implementation for `julienschmidt/httprouter`.
type FastRouter struct {
	mux *httprouter.Router
}

// Handle will call the `httprouter.METHOD` methods and use the HTTPToFastRoute
// to pass httprouter.Params into a Gorilla request context. The params will be available
// via the `FastRouterVars` function.
func (g *FastRouter) Handle(method, path string, h http.Handler) {
	switch strings.ToUpper(method) {
	case "GET":
		g.mux.GET(path, HTTPToFastRoute(h))
	case "PUT":
		g.mux.PUT(path, HTTPToFastRoute(h))
	case "POST":
		g.mux.POST(path, HTTPToFastRoute(h))
	case "DELETE":
		g.mux.DELETE(path, HTTPToFastRoute(h))
	case "HEAD":
		g.mux.HEAD(path, HTTPToFastRoute(h))
	case "OPTIONS":
		g.mux.OPTIONS(path, HTTPToFastRoute(h))
	case "PATCH":
		g.mux.PATCH(path, HTTPToFastRoute(h))
	default:
		g.mux.Handle(method, path, HTTPToFastRoute(h))
	}
}

// HandleFunc will call the `httprouter.METHOD` methods and use the HTTPToFastRoute
// to pass httprouter.Params into a Gorilla request context. The params will be available
// via the `FastRouterVars` function.
func (g *FastRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will set httprouter.Router.NotFound.
func (g *FastRouter) SetNotFoundHandler(h http.Handler) {
	g.mux.NotFound = h
}

// ServeHTTP will call httprouter.ServerHTTP directly.
func (g *FastRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// HTTPToFastRoute will convert an http.Handler to a httprouter.Handle
// by stuffing any route parameters into a Gorilla request context.
// To access the request parameters within the endpoint,
// use the `FastRouterVars` function.
func HTTPToFastRoute(fh http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if len(params) > 0 {
			vars := map[string]string{}
			for _, param := range params {
				vars[param.Key] = param.Value
			}
			setFastRouteVars(r, vars)
		}
		fh.ServeHTTP(w, r)
	}
}

// FastRouteVars is a helper function for accessing route
// parameters from the FastRouter. This is the equivalent
// of using `mux.Vars(r)` with the GorillaRouter.
func FastRouteVars(r *http.Request) map[string]string {
	if rv := context.Get(r, fastRouteVarsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

func setFastRouteVars(r *http.Request, val interface{}) {
	if val != nil {
		context.Set(r, fastRouteVarsKey, val)
	}
}
