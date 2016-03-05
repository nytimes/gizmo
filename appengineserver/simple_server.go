package appengineserver

import (
	"errors"
	"net/http"

	"github.com/NYTimes/gizmo/config"
	"github.com/gorilla/mux"
)

// SimpleServer is a basic http Server implementation for
// serving SimpleService, JSONService or MixedService implementations.
type SimpleServer struct {
	cfg *config.Server

	// exit chan for graceful shutdown
	exit chan chan error

	// mux for routing
	mux *mux.Router
}

// NewSimpleServer will init the mux, exit channel and
// build the address from the given port. It will register the HealthCheckHandler
// at the given path and set up the shutDownHandler to be called on Stop().
func NewSimpleServer(cfg *config.Server) *SimpleServer {
	if cfg == nil {
		cfg = &config.Server{}
	}
	mx := mux.NewRouter()
	if cfg.NotFoundHandler != nil {
		mx.NotFoundHandler = cfg.NotFoundHandler
	}
	return &SimpleServer{
		mux: mx,
		cfg: cfg,
	}
}

// ServeHTTP is SimpleServer's hook for metrics and safely executing each request.
func (s *SimpleServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hand the request off to gorilla
	s.mux.ServeHTTP(w, r)
}

// UnexpectedServerError is returned with a 500 status code when SimpleServer recovers
// from a panic in a request.
var UnexpectedServerError = []byte("unexpected server error")

// Register will accept and register SimpleServer, JSONService or MixedService implementations.
func (s *SimpleServer) Register(svcI Service) error {
	var (
		js JSONService
		ss SimpleService
	)
	switch svc := svcI.(type) {
	case MixedService:
		js = svc
		ss = svc
	case SimpleService:
		ss = svc
	case JSONService:
		js = svc
	default:
		return errors.New("services for SimpleServers must implement the SimpleService, JSONService or MixedService interfaces")
	}

	sr := s.mux.PathPrefix(svcI.Prefix()).Subrouter()

	if ss != nil {
		// register all simple endpoints with our wrappers
		for path, epMethods := range ss.Endpoints() {
			for method, ep := range epMethods {
				sr.Handle(path,
					ss.Middleware(
						ContextToHTTP(ss.ContextMiddleware(ep)),
					),
				).Methods(method)
			}
		}
	}

	if js != nil {
		// register all JSON endpoints with our wrapper
		for path, epMethods := range js.JSONEndpoints() {
			for method, ep := range epMethods {
				sr.Handle(path,
					js.Middleware(
						ContextToHTTP(
							js.ContextMiddleware(JSONToHTTPContext(js.JSONMiddleware(ep))),
						),
					),
				).Methods(method)
			}
		}
	}

	return nil
}
