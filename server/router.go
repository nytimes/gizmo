package server

import (
	"net/http"

	"github.com/gorilla/mux"
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
	case "gorilla":
		return &GorillaRouter{mux.NewRouter()}
	case "stdlib":
		return &stdlibRouter{mux: http.NewServeMux()}
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
		SetRouteVars(r, mux.Vars(r))
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

// stdlibRouter is a Router implementation for the stdlib's `http.ServeMux`.
type stdlibRouter struct {
	mux *http.ServeMux
}

// Handle will call the Stdlib's HandleFunc() methods with a check for the incoming
// HTTP method. To allow for multiple methods on a single route, use 'ANY'.
func (g *stdlibRouter) Handle(method, path string, h http.Handler) {
	g.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method || method == "ANY" {
			h.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

// HandleFunc will call the Stdlib's HandleFunc() methods with a check for the incoming
// HTTP method. To allow for multiple methods on a single route, use 'ANY'.
func (g *stdlibRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will do nothing as we cannot override the stdlib not found.
func (g *stdlibRouter) SetNotFoundHandler(h http.Handler) {
}

// ServeHTTP will call Stdlib's ServeMux.ServerHTTP directly.
func (g *stdlibRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}
