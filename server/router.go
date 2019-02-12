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
