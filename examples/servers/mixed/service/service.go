package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gziphandler"
	"github.com/sirupsen/logrus"

	"github.com/NYTimes/gizmo/examples/nyt"
)

type (
	// MixedService will implement server.MixedService and
	// handle all requests to the server.
	MixedService struct {
		client nyt.Client
	}
	// Config is a struct to contain all the needed
	// configuration for our MixedService
	Config struct {
		Server           *server.Config
		MostPopularToken string
		SemanticToken    string
	}
)

// NewMixedService will instantiate a MixedService
// with the given configuration.
func NewMixedService(cfg *Config) *MixedService {
	return &MixedService{
		nyt.NewClient(cfg.MostPopularToken, cfg.SemanticToken),
	}
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *MixedService) Prefix() string {
	return "/svc/nyt"
}

// Middleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we're using a GzipHandler middleware to
// compress our responses.
func (s *MixedService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

// JSONMiddleware provides a JSONEndpoint hook wrapped around all requests.
// In this implementation, we're using it to provide application logging and to check errors
// and provide generic responses.
func (s *MixedService) JSONMiddleware(j server.JSONEndpoint) server.JSONEndpoint {
	return func(r *http.Request) (int, interface{}, error) {

		status, res, err := j(r)
		if err != nil {
			server.LogWithFields(r).WithFields(logrus.Fields{
				"error": err,
			}).Error("problems with serving request")
			return http.StatusServiceUnavailable, nil, &jsonErr{"sorry, this service is unavailable"}
		}

		server.LogWithFields(r).Info("success!")
		return status, res, nil
	}
}

// Endpoints is a listing of all endpoints available in the MixedService.
func (s *MixedService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/cats": map[string]http.HandlerFunc{
			"GET": s.GetCats,
		},
	}
}

// JSONEndpoints is a listing of all JSON endpoints available in the MixedService.
func (s *MixedService) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	return map[string]map[string]server.JSONEndpoint{
		"/most-popular/{resourceType}/{section}/{timeframe}": map[string]server.JSONEndpoint{
			"GET": s.GetMostPopular,
		},
	}
}
