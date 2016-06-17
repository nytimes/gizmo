/*
Package server is the bulk of the toolkit and relies on `server.Config` for any managing `Server` implementations. A server must implement the following interface:

    // Server is the basic interface that defines what expect from any server.
    type Server interface {
        Register(...Service) error
        Start() error
        Stop() error
    }

The package offers 2 server implementations:

`SimpleServer`, which is capable of handling basic HTTP and JSON requests via 3 of the available `Service` implementations: `SimpleService`, `JSONService`, `ContextService`, `MixedService` and `MixedContextService`. A service and these implementations will be defined below.

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

        // Ensure that the route syntax is compatible with the router
        // implementation chosen in cfg.RouterType.
        // route - method - func
        JSONEndpoints() map[string]map[string]JSONEndpoint
        // JSONMiddleware provides a hook for service-wide middleware around JSONEndpoints.
        JSONMiddleware(JSONEndpoint) JSONEndpoint
    }

    type MixedService interface {
        Service

        // route - method - func
        Endpoints() map[string]map[string]http.HandlerFunc

        // Ensure that the route syntax is compatible with the router
        // implementation chosen in cfg.RouterType.
        // route - method - func
        JSONEndpoints() map[string]map[string]JSONEndpoint
        // JSONMiddleware provides a hook for service-wide middleware around JSONEndpoints.
        JSONMiddleware(JSONEndpoint) JSONEndpoint
    }

    type ContextService interface {
        Service

        // route - method - func
        ContextEndpoints() map[string]map[string]ContextHandlerFunc
        // ContextMiddleware provides a hook for service-wide middleware around ContextHandler
        ContextMiddleware(ContextHandler) ContextHandler
    }

    type MixedContextService interface {
        ContextService

        // route - method - func
        JSONEndpoints() map[string]map[string]JSONContextEndpoint
        JSONContextMiddleware(JSONContextEndpoint) JSONContextEndpoint
    }

Where `JSONEndpoint`, `JSONContextEndpoint`, `ContextHandler` and `ContextHandlerFunc` are defined as:

    type JSONEndpoint func(*http.Request) (int, interface{}, error)

    type JSONContextEndpoint func(context.Context, *http.Request) (int, interface{}, error)

    type ContextHandler interface {
        ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
    }

    type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

Also, the one service type that works with an `RPCServer`:

    type RPCService interface {
        ContextService

        Service() (grpc.ServiceDesc, interface{})

        // Ensure that the route syntax is compatible with the router
        // implementation chosen in cfg.RouterType.
        // route - method - func
        JSONEndpoints() map[string]map[string]JSONContextEndpoint
        // JSONMiddleware provides a hook for service-wide middleware around JSONContextEndpoints.
        JSONMiddlware(JSONContextEndpoint) JSONContextEndpoint
    }

The `Middleware(..)` functions offer each service a 'hook' to wrap each of its endpoints. This may be handy for adding additional headers or context to the request. This is also the point where other, third-party middleware could be easily be plugged in (ie. oauth, tracing, metrics, logging, etc.)

Examples

Check out the gizmo/examples/servers directory to see several reference implementations.
*/
package server
