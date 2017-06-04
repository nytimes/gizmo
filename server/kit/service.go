package kit

import (
	"net/http"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Endpoint encapsulates everything required to build
// an endpoint hosted on an HTTP server.
type Endpoint struct {
	Endpoint    endpoint.Endpoint
	HTTPDecoder httptransport.DecodeRequestFunc
	HTTPEncoder httptransport.EncodeResponseFunc
	HTTPOptions []httptransport.ServerOption
}

// Service is the most basic interface of a service that can be received and
// hosted by a kithttp server. Services provide hooks for service-wide options and
// middlewares and can be used as a means of dependency injection.
// In general, a Service should just contain the logic for deserializing HTTP
// requests, passing the request to a business logic interface abstraction,
// handling errors and serializing the apprioriate response.
//
// In other words, each Endpoint is similar to a 'controller' and the Service
// a container for injecting depedencies (business services, repositories, etc.)
// into each request handler.
//
type Service interface {
	// HTTPMiddleware is for service-wide http specific middlewares
	// for easy integration with 3rd party http.Handlers.
	HTTPMiddleware(http.Handler) http.Handler

	// Middleware is for any service-wide go-kit middlewares
	Middleware(endpoint.Endpoint) endpoint.Endpoint

	// Options are service-wide go-kit options
	Options() []httptransport.ServerOption

	// RouterOptions allows users to override the default
	// behavior and use of the GorillaRouter.
	RouterOptions() []RouterOption

	// HTTPEndpoints default to using a JSON serializer if no encoder is provided.
	// For example:
	//
	//    return map[string]map[string]kithttp.HTTPEndpoint{
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
	HTTPEndpoints() map[string]map[string]Endpoint

	// in case you want to run this as a gRPC server too
	ServiceDesc() *grpc.ServiceDesc
	RPCOptions() []grpc.ServerOption
}
