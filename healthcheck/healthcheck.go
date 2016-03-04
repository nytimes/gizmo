package healthcheck

import (
	"io"
	"net/http"

	"github.com/NYTimes/gizmo/config"
)

// NewHealthCheckHandler will inspect the config to generate
// the appropriate HealthCheckHandler.
func NewHandler(cfg *config.Server) Handler {
	switch cfg.HealthCheckType {
	case "simple":
		return NewSimple(cfg.HealthCheckPath)
	case "esx":
		return NewESX()
	case "appengine":
		return NewSimple("/_ah/health")
	default:
		return NewSimple("/status.txt")
	}
}

// Handler is an interface used by SimpleServer and RPCServer
// to allow users to customize their service's health check. Start will be
// called just before server start up and the given ActivityMonitor should
// offer insite to the # of requests in flight, if needed.
// Stop will be called once the servers receive a kill signal.
type Handler interface {
	http.Handler
	Path() string
	Start(*ActivityMonitor) error
	Stop() error
}

// Simple is a basic Handler implementation
// that _always_ returns with an "ok" status and shuts down immediately.
type Simple struct {
	path string
}

// NewSimple will return a new Simple instance.
func NewSimple(path string) *Simple {
	return &Simple{path: path}
}

// Path will return the configured status path to server on.
func (s *Simple) Path() string {
	return s.path
}

// Start will do nothing.
func (s *Simple) Start(monitor *ActivityMonitor) error {
	return nil
}

// Stop will do nothing and return nil.
func (s *Simple) Stop() error {
	return nil
}

// ServeHTTP will always respond with "ok-"+server.Name.
func (s *Simple) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}
