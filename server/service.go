package server

import (
	"net/http"

	"google.golang.org/grpc"
)

// Service is the most basic interface of a service that can be received and
// hosted by a Server.
type Service interface {
	Prefix() string

	// Middleware is a hook to enable services to add
	// any additional middleware
	Middleware(http.Handler) http.Handler
}

// SimpleService is an interface defining a service that
// is made up of http.HandlerFuncs
type SimpleService interface {
	Service

	// route - method - func
	Endpoints() map[string]map[string]http.HandlerFunc
}

// JSONService is an interface defining a service that
// is made up of JSONEndpoints.
type JSONService interface {
	Service

	// route - method - func
	JSONEndpoints() map[string]map[string]JSONEndpoint
	JSONMiddleware(JSONEndpoint) JSONEndpoint
}

// MixedService is an interface defining service that
// offer JSONEndpoints and simple http.HandlerFunc endpoints.
type MixedService interface {
	Service

	// route - method - func
	Endpoints() map[string]map[string]http.HandlerFunc

	// route - method - func
	JSONEndpoints() map[string]map[string]JSONEndpoint
	JSONMiddleware(JSONEndpoint) JSONEndpoint
}

// RPCService is an interface defining an grpc-compatible service that
// offers JSONEndpoints.
type RPCService interface {
	Service

	Service() (*grpc.ServiceDesc, interface{})

	// route - method - func
	JSONEndpoints() map[string]map[string]JSONEndpoint
	JSONMiddleware(JSONEndpoint) JSONEndpoint
}

// JSONEndpoint is the JSONService equivalent to SimpleService's http.HandlerFunc
type JSONEndpoint func(*http.Request) (int, interface{}, error)
