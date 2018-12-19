package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gziphandler"
	"github.com/sirupsen/logrus"
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
func NewJSONPubService(cfg aws.SNSConfig) *JSONPubService {
	pub, err := aws.NewPublisher(cfg)
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
