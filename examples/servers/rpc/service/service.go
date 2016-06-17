package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gziphandler"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/NYTimes/gizmo/examples/nyt"
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
		Server           *server.Config
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

// ContextMiddleware provides a server.ContextHAndler hook wrapped around all
// requests. This could be handy if you need to decorate the request context.
func (s *RPCService) ContextMiddleware(h server.ContextHandler) server.ContextHandler {
	return h
}

// JSONMiddleware provides a JSONContextEndpoint hook wrapped around all requests.
// In this implementation, we're using it to provide application logging and to check errors
// and provide generic responses.
func (s *RPCService) JSONMiddleware(j server.JSONContextEndpoint) server.JSONContextEndpoint {
	return func(ctx context.Context, r *http.Request) (int, interface{}, error) {
		status, res, err := j(ctx, r)
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

// ContextEndpoints may be needed if your server has any non-RPC-able
// endpoints. In this case, we have none but still need this method to
// satisfy the server.RPCService interface.
func (s *RPCService) ContextEndpoints() map[string]map[string]server.ContextHandlerFunc {
	return map[string]map[string]server.ContextHandlerFunc{}
}

// JSONContextEndpoints is a listing of all endpoints available in the RPCService.
func (s *RPCService) JSONEndpoints() map[string]map[string]server.JSONContextEndpoint {
	return map[string]map[string]server.JSONContextEndpoint{
		"/most-popular/{resourceType}/{section}/{timeframe}": map[string]server.JSONContextEndpoint{
			"GET": s.GetMostPopularJSON,
		},
		"/cats": map[string]server.JSONContextEndpoint{
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
