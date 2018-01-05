package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"

	"github.com/NYTimes/gizmo/web"
)

// Router is an interface to wrap different router types to be embedded within
// Gizmo server.Server implementations.
type Router interface {
	Handle(method string, path string, handler http.Handler)
	HandleFunc(method string, path string, handlerFunc func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	SetNotFoundHandler(handler http.Handler)
}

// NewRouter will return the router specified by the server
// config. If no Router value is supplied, the server
// will default to using Gorilla mux.
func NewRouter(cfg *Config) Router {
	switch cfg.RouterType {
	case "fast", "httprouter":
		return &FastRouter{httprouter.New()}
	case "gorilla":
		return &GorillaRouter{mux.NewRouter()}
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
	g.mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// copy the route params into a shared location
		// duplicating memory, but allowing Gizmo to be more flexible with
		// router implementations.
		web.SetRouteVars(r, mux.Vars(r))
		h.ServeHTTP(w, r)
	})).Methods(method)
}

// HandleFunc will call the Gorilla web toolkit's HandleFunc().Method() methods.
func (g *GorillaRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will set the Gorilla mux.Router.NotFoundHandler.
func (g *GorillaRouter) SetNotFoundHandler(h http.Handler) {
	g.mux.NotFoundHandler = h
}

// ServeHTTP will call Gorilla mux.Router.ServerHTTP directly.
func (g *GorillaRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// FastRouter is a Router implementation for `julienschmidt/httprouter`. THIS ROUTER IS
// DEPRECATED. Please use gorilla or stdlib if you can. Metrics will not work properly on
// servers using this router type. (see issue #132)
type FastRouter struct {
	mux *httprouter.Router
}

// Handle will call the `httprouter.METHOD` methods and use the HTTPToFastRoute
// to pass httprouter.Params into a Gorilla request context. The params will be available
// via the `FastRouterVars` function.
func (f *FastRouter) Handle(method, path string, h http.Handler) {
	f.mux.Handle(method, path, HTTPToFastRoute(h))
}

// HandleFunc will call the `httprouter.METHOD` methods and use the HTTPToFastRoute
// to pass httprouter.Params into a Gorilla request context. The params will be available
// via the `FastRouterVars` function.
func (f *FastRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	f.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will set httprouter.Router.NotFound.
func (f *FastRouter) SetNotFoundHandler(h http.Handler) {
	f.mux.NotFound = h
}

// ServeHTTP will call httprouter.ServerHTTP directly.
func (f *FastRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mux.ServeHTTP(w, r)
}

// HTTPToFastRoute will convert an http.Handler to a httprouter.Handle
// by stuffing any route parameters into a Gorilla request context.
// To access the request parameters within the endpoint,
// use the `web.Vars` function.
func HTTPToFastRoute(fh http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if len(params) > 0 {
			vars := map[string]string{}
			for _, param := range params {
				vars[param.Key] = param.Value
			}
			web.SetRouteVars(r, vars)
		}
		fh.ServeHTTP(w, r)
	}
}
