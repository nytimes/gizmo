package kit

import (
	"net/http"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// HTTPEndpoint encapsulates everything required to build
// an endpoint hosted on a kit server.
type HTTPEndpoint struct {
	Endpoint endpoint.Endpoint
	Decoder  httptransport.DecodeRequestFunc
	Encoder  httptransport.EncodeResponseFunc
	Options  []httptransport.ServerOption
}

// Service is the interface of mixed HTTP/gRPC that can be registered and
// hosted by a gizmo/server/kit server. Services provide hooks for service-wide options
// and middlewares and can be used as a means of dependency injection.
// In general, a Service should just contain the logic for deserializing and authorizing
// requests, passing the request to a business logic interface abstraction,
// handling errors and serializing the apprioriate response.
//
// In other words, each Endpoint is similar to a 'controller' and the Service
// a container for injecting dependencies (business services, repositories, etc.)
// into each request handler.
type Service interface {
	// Middleware is for any service-wide go-kit middlewares. This middleware
	// is applied to both HTTP and gRPC services.
	Middleware(endpoint.Endpoint) endpoint.Endpoint

	// HTTPMiddleware is for service-wide http specific middleware
	// for easy integration with 3rd party http.Handlers like nytimes/gziphandler.
	HTTPMiddleware(http.Handler) http.Handler

	// HTTPOptions are service-wide go-kit HTTP server options
	HTTPOptions() []httptransport.ServerOption

	// HTTPRouterOptions allows users to override the default
	// behavior and use of the GorillaRouter.
	HTTPRouterOptions() []RouterOption

	// HTTPEndpoints default to using a JSON serializer if no encoder is provided.
	// For example:
	//
	//    return map[string]map[string]kit.HTTPEndpoint{
	//        "/cat/{id}": {
	//            "GET": {
	//                Endpoint: s.GetCatByID,
	//                Decoder:  decodeGetCatRequest,
	//            },
	//        },
	//        "/cats": {
	//            "PUT": {
	//                Endpoint: s.PutCats,
	//                HTTPDecoder:  decodePutCatsProtoRequest,
	//            },
	//            "GET": {
	//                Endpoint: s.GetCats,
	//                HTTPDecoder:  decodeGetCatsRequest,
	//            },
	//        },
	//  }
	HTTPEndpoints() map[string]map[string]HTTPEndpoint

	// RPCMiddleware is for any service-wide gRPC specific middleware
	// for easy integration with 3rd party grpc.UnaryServerInterceptors like
	// http://godoc.org/cloud.google.com/go/trace#Client.GRPCServerInterceptor
	//
	// The underlying kit server already uses the one available grpc.UnaryInterceptor
	// grpc.ServerOption so attempting to pass your own in this Service's RPCOptions()
	// will cause a panic at startup.
	//
	// If you want to apply multiple RPC middlewares,
	// we recommend using:
	// http://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware#ChainUnaryServer
	RPCMiddleware() grpc.UnaryServerInterceptor

	// RPCServiceDesc allows services to declare an alternate gRPC
	// representation of themselves to be hosted on the RPC_PORT (8081 by default).
	RPCServiceDesc() *grpc.ServiceDesc

	// RPCOptions are for service-wide gRPC server options.
	//
	// The underlying kit server already uses the one available grpc.UnaryInterceptor
	// grpc.ServerOption so attempting to pass your own in this method will cause a panic
	// at startup. We recommend using RPCMiddleware() to fill this need.
	RPCOptions() []grpc.ServerOption
}

// Shutdowner allows your service to shutdown gracefully when http server stops.
// This may used when service has any background task which needs to be completed gracefully.
type Shutdowner interface {
	Shutdown()
}
