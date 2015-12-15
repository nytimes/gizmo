package service

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/Sirupsen/logrus"
	"github.com/nytimes/gizmo/config"
	"github.com/nytimes/gizmo/pubsub"
	"github.com/nytimes/gizmo/server"
)

type (
	// JSONPubService will implement server.JSONPubService and
	// handle all requests to the server.
	JSONPubService struct {
		pub pubsub.Publisher
	}
)

// NewJSONPubService will instantiate a JSONPubService
// with the given configuration.
func NewJSONPubService(cfg *config.Config) *JSONPubService {
	pub, err := pubsub.NewSNSPublisher(cfg.SNS)
	if err != nil {
		server.Log.Fatal("unable to init publisher: ", err)
	}
	return &JSONPubService{pub}
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *JSONPubService) Prefix() string {
	return "/svc/nyt"
}

// Middleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we're using a GzipHandler middleware to
// compress our responses.
func (s *JSONPubService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

// JSONMiddleware provides a JSONEndpoint hook wrapped around all requests.
// In this implementation, we're using it to provide application logging and to check errors
// and provide generic responses.
func (s *JSONPubService) JSONMiddleware(j server.JSONEndpoint) server.JSONEndpoint {
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

// JSONEndpoints is a listing of all endpoints available in the JSONPubService.
func (s *JSONPubService) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	return map[string]map[string]server.JSONEndpoint{
		"/cats": map[string]server.JSONEndpoint{
			"PUT": s.PublishCats,
		},
	}
}

type jsonErr struct {
	Err string `json:"error"`
}

func (e *jsonErr) Error() string {
	return e.Err
}
