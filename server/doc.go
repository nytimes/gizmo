/*
Package server is the bulk of the toolkit and relies on `config` for any managing `Server` implementations. A server must implement the following interface:

    // Server is the basic interface that defines what expect from any server.
    type Server interface {
        Register(...Service) error
        Start() error
        Stop() error
    }

The package offers 2 server implementations:

`SimpleServer`, which is capable of handling basic HTTP and JSON requests via 3 of the available `Service` implementations: `SimpleService`, `JSONService`, and `MixedService`. A service and these implementations will be defined below.

`RPCServer`, which is capable of serving a gRPC server on one port and JSON endpoints on another. This kind of server can only handle the `RPCService` implementation.

The `Service` interface is minimal to allow for maximum flexibility:

    type Service interface {
        Prefix() string

        // Middleware provides a hook for service-wide middleware
        Middleware(http.Handler) http.Handler
    }

The 3 service types that are accepted and hostable on the `SimpleServer`:

    type SimpleService interface {
        Service

        // router - method - func
        Endpoints() map[string]map[string]http.HandlerFunc
    }

    type JSONService interface {
        Service

        // router - method - func
        JSONEndpoints() map[string]map[string]JSONEndpoint
        // JSONMiddleware provides a hook for service-wide middleware around JSONEndpoints.
        JSONMiddleware(JSONEndpoint) JSONEndpoint
    }

    type MixedService interface {
        Service

        // route - method - func
        Endpoints() map[string]map[string]http.HandlerFunc

        // route - method - func
        JSONEndpoints() map[string]map[string]JSONEndpoint
        // JSONMiddleware provides a hook for service-wide middleware around JSONEndpoints.
        JSONMiddleware(JSONEndpoint) JSONEndpoint
    }

Where a `JSONEndpoint` is defined as:

    type JSONEndpoint func(*http.Request) (int, interface{}, error)

Also, the one service type that works with an `RPCServer`:

    type RPCService interface {
        Service

        Service() (grpc.ServiceDesc, interface{})

        // route - method - func
        JSONEndpoints() map[string]map[string]JSONEndpoint
        // JSONMiddleware provides a hook for service-wide middleware around JSONEndpoints.
        JSONMiddlware(JSONEndpoint) JSONEndpoint
    }

The `Middleware(..)` functions offer each service a 'hook' to wrap each of its endpoints. This may be handy for adding additional headers or context to the request. This is also the point where other, third-party middleware could be easily be plugged in (ie. oauth, tracing, metrics, logging, etc.)

Examples

Check out the gizmo/examples/servers directory to see several reference implementations.
*/
package server
