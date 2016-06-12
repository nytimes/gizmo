package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/pubsub/kafka"
	"github.com/NYTimes/gizmo/server"
)

// StreamService offers three endpoints: one to create a new topic in
// Kafka, a second to expose the topic over a websocket and a third
// to host a web page that provides a demo.
type StreamService struct {
	port int
	cfg  *kafka.Config
}

// NewStreamService will return a new stream service instance.
// If the given config is empty, it will default to localhost.
func NewStreamService(port int, cfg *kafka.Config) *StreamService {
	if cfg == nil {
		cfg = &kafka.Config{BrokerHosts: []string{"localhost:9092"}}
	}
	return &StreamService{port, cfg}
}

// Prefix is the string prefixed to all endpoint routes.
func (s *StreamService) Prefix() string {
	return "/svc/v1"
}

// Middleware in this service will do nothing.
func (s *StreamService) Middleware(h http.Handler) http.Handler {
	return server.NoCacheHandler(h)
}

// Endpoints returns the two endpoints for our stream service.
func (s *StreamService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/create": map[string]http.HandlerFunc{
			"GET": server.JSONToHTTP(s.CreateStream).ServeHTTP,
		},
		"/stream/{stream_id:[0-9]+}": map[string]http.HandlerFunc{
			"GET": s.Stream,
		},
		"/demo/{stream_id:[0-9]+}": map[string]http.HandlerFunc{
			"GET": s.Demo,
		},
	}
}
