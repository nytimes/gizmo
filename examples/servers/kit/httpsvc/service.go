package httpsvc

import (
	"net/http"

	"github.com/NYTimes/gizmo/server/kithttp"
	"github.com/NYTimes/gziphandler"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/NYTimes/gizmo/examples/nyt"
)

type (
	// httpService will implement kithttp.JSONService.
	httpService struct {
		client nyt.Client
	}
	// Config is a struct to contain all the needed
	// configuration for our JSONService
	Config struct {
		MostPopularToken string `envconfig:"MOST_POPULAR_TOKEN"`
		SemanticToken    string `envconfig:"SEMANTIC_TOKEN"`
	}
)

// NewJSONService will instantiate a JSONService
// with the given configuration.
func New(cfg Config) kithttp.JSONService {
	return httpService{
		nyt.NewClient(cfg.MostPopularToken, cfg.SemanticToken),
	}
}

func (s httpService) RouterOptions() []kithttp.RouterOption {
	return nil
}

func (s httpService) Options() []httptransport.ServerOption {
	return nil
}

// HTTPMiddleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we're using a GzipHandler middleware to
// compress our responses.
func (s httpService) HTTPMiddleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

// Middleware provides an kit/endpoint.Middleware hook wrapped around all requests.
func (s httpService) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return e
}

// JSONEndpoints is a listing of all endpoints available in the JSONService.
func (s httpService) JSONEndpoints() map[string]map[string]kithttp.HTTPEndpoint {
	return map[string]map[string]kithttp.HTTPEndpoint{
		"/most-popular/{resourceType:[a-z]+}/{section:[a-z]+}/{timeframe:[0-9]+}": {
			"GET": {
				Endpoint: s.GetMostPopular,
				Decoder:  decodeMostPopularRequest,
			},
		},
		"/cats": {
			"GET": {
				Endpoint: s.GetCats,
			},
		},
	}
}
