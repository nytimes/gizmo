package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gziphandler"

	"github.com/NYTimes/gizmo/examples/nyt"
)

type (
	// SimpleService will implement server.SimpleService and
	// handle all requests to the server.
	SimpleService struct {
		client nyt.Client
	}
	// Config is a struct to contain all the needed
	// configuration for our SimpleService
	Config struct {
		Server           *server.Config
		MostPopularToken string `envconfig:"MOST_POPULAR_TOKEN"`
		SemanticToken    string `envconfig:"SEMANTIC_TOKEN"`
	}
)

// NewSimpleService will instantiate a SimpleService
// with the given configuration.
func NewSimpleService(cfg *Config) *SimpleService {
	return &SimpleService{
		nyt.NewClient(cfg.MostPopularToken, cfg.SemanticToken),
	}
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *SimpleService) Prefix() string {
	return "/svc/nyt"
}

// Middleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we're using a GzipHandler middleware to
// compress our responses.
func (s *SimpleService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

// JSONEndpoints is a listing of all endpoints available in the SimpleService.
func (s *SimpleService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/most-popular/{resourceType}/{section}/{timeframe}": map[string]http.HandlerFunc{
			"GET": server.JSONToHTTP(s.GetMostPopular).ServeHTTP,
		},
		"/cats": map[string]http.HandlerFunc{
			"GET": s.GetCats,
		},
	}
}
