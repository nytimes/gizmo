package appengine

import (
	"net/http"

	"github.com/NYTimes/gizmo/appengineserver"
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gziphandler"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"

	"github.com/NYTimes/gizmo/examples/nyt"
)

type (
	// AppEngineService will implement appengineserver.JSONService and
	// handle all requests to the server.
	AppEngineService struct {
		client nyt.ContextClient
	}
	// Config is a struct to contain all the needed
	// configuration for our AppEngineService
	Config struct {
		*config.Server
		MostPopularToken string `envconfig:"MOST_POPULAR_TOKEN"`
		SemanticToken    string `envconfig:"SEMANTIC_TOKEN"`
	}
)

func init() {
	var cfg Config
	config.LoadEnvConfig(&cfg)
	appengineserver.Init(cfg.Server, NewAppEngineService(&cfg))
}

// NewAppEngineService will instantiate a AppEngineService
// with the given configuration.
func NewAppEngineService(cfg *Config) *AppEngineService {
	return &AppEngineService{
		nyt.NewContextClient(cfg.MostPopularToken, cfg.SemanticToken),
	}
}

// need to find a way to preload this? maybe a 'warmup'?
func (s *AppEngineService) nytclient() nyt.ContextClient {
	var cfg Config
	config.LoadEnvConfig(&cfg)
	s.client = nyt.NewContextClient(cfg.MostPopularToken, cfg.SemanticToken)
	return s.client
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *AppEngineService) Prefix() string {
	return "/svc/nyt"
}

// Middleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we're using a GzipHandler middleware to
// compress our responses.
func (s *AppEngineService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

func (s *AppEngineService) ContextMiddleware(h appengineserver.ContextHandler) appengineserver.ContextHandler {
	return h
}

// JSONMiddleware provides a JSONEndpoint hook wrapped around all requests.
// In this implementation, we're using it to provide application logging and to check errors
// and provide generic responses.
func (s *AppEngineService) JSONMiddleware(j appengineserver.JSONEndpoint) appengineserver.JSONEndpoint {
	return func(ctx context.Context, r *http.Request) (int, interface{}, error) {

		status, res, err := j(ctx, r)
		if err != nil {
			log.Warningf(ctx, "problems with the request: %s", err)
			return http.StatusServiceUnavailable, nil, &jsonErr{"sorry, this service is unavailable"}
		}

		return status, res, nil
	}
}

// JSONEndpoints is a listing of all endpoints available in the AppEngineService.
func (s *AppEngineService) JSONEndpoints() map[string]map[string]appengineserver.JSONEndpoint {
	return map[string]map[string]appengineserver.JSONEndpoint{
		"/most-popular/{resourceType}/{section}/{timeframe}": map[string]appengineserver.JSONEndpoint{
			"GET": s.GetMostPopular,
		},
		"/cats": map[string]appengineserver.JSONEndpoint{
			"GET": s.GetCats,
		},
	}
}

type jsonErr struct {
	Err string `json:"error"`
}

func (e *jsonErr) Error() string {
	return e.Err
}
