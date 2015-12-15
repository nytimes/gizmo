package service

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/Sirupsen/logrus"
	"github.com/nytimes/gizmo/config"
	"github.com/nytimes/gizmo/server"
	"google.golang.org/grpc"

	"github.com/nytimes/gizmo/examples/nyt"
)

type (
	// RPCService will implement server.RPCService and
	// handle all requests to the server.
	RPCService struct {
		client nyt.Client
	}
	// Config is a struct to contain all the needed
	// configuration for our RPCService
	Config struct {
		*config.Server
		MostPopularToken string
		SemanticToken    string
	}
)

// NewRPCService will instantiate a RPCService
// with the given configuration.
func NewRPCService(cfg *Config) *RPCService {
	return &RPCService{
		nyt.NewClient(cfg.MostPopularToken, cfg.SemanticToken),
	}
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *RPCService) Prefix() string {
	return "/svc/nyt"
}

// Service provides the RPCService with a description of the
// service to serve and the implementation.
func (s *RPCService) Service() (*grpc.ServiceDesc, interface{}) {
	return &NYTProxyService_serviceDesc, s
}

// Middleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we're using a GzipHandler middleware to
// compress our responses.
func (s *RPCService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

// JSONMiddleware provides a JSONEndpoint hook wrapped around all requests.
// In this implementation, we're using it to provide application logging and to check errors
// and provide generic responses.
func (s *RPCService) JSONMiddleware(j server.JSONEndpoint) server.JSONEndpoint {
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

// JSONEndpoints is a listing of all endpoints available in the RPCService.
func (s *RPCService) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	return map[string]map[string]server.JSONEndpoint{
		"/most-popular/{resourceType}/{section}/{timeframe}": map[string]server.JSONEndpoint{
			"GET": s.GetMostPopularJSON,
		},
		"/cats": map[string]server.JSONEndpoint{
			"GET": s.GetCatsJSON,
		},
	}
}

type jsonErr struct {
	Err string `json:"error"`
}

func (e *jsonErr) Error() string {
	return e.Err
}
