# Gizmo Microservice Toolkit [![GoDoc](https://godoc.org/github.com/gizmo/gizmo?status.svg)](https://godoc.org/github.com/NYTimes/gizmo) [![Build Status](https://travis-ci.org/NYTimes/gizmo.svg?branch=master)](https://travis-ci.org/NYTimes/gizmo) [![Coverage Status](https://coveralls.io/repos/NYTimes/gizmo/badge.svg?branch=master&service=github)](https://coveralls.io/github/NYTimes/gizmo?branch=master)

![Gizmo!](http://graphics8.nytimes.com/images/blogs/open/2015/gizmo.png)

This toolkit provides packages to put together server and pubsub daemons with the following features:

* standardized configuration and logging
* health check endpoints with configurable strategies
* configuration for managing pprof endpoints and log levels
* structured logging containing basic request information
* useful metrics for endpoints
* graceful shutdowns
* basic interfaces to define our expectations and vocabulary

In this toolkit, you will find:

## The `config` packages

The `config` package contains a handful of useful functions to load to configuration structs from JSON files, JSON blobs in Consul k/v or environment variables.

The subpackages contain structs meant for managing common configuration options and credentials. There are currently configs for:

* Go Kit Metrics
* MySQL
* MongoDB
* Oracle
* PostgreSQL
* AWS (S3, DynamoDB, ElastiCache)
* GCP
* Gorilla's `securecookie`

The package also has a generic `Config` type in the `config/combined` subpackage that contains all of the above types. It's meant to be a 'catch all' convenience struct that many applications should be able to use.

## The `server` package

This package is the bulk of the toolkit and relies on `server.Config` for any managing `Server` implementations. A server must implement the following interface:

```go
// Server is the basic interface that defines what expect from any server.
type Server interface {
    Register(Service) error
    Start() error
    Stop() error
}
```

The package offers 2 server implementations:

`SimpleServer`, which is capable of handling basic HTTP and JSON requests via 5 of the available `Service` implementations: `SimpleService`, `JSONService`, `ContextService`, `MixedService` and a `MixedContextService`. A service and these implementations will be defined below.

`RPCServer`, which is capable of serving a gRPC server on one port and JSON endpoints on another. This kind of server can only handle the `RPCService` implementation.

The `Service` interface is minimal to allow for maximum flexibility:
```go
type Service interface {
    Prefix() string

    // Middleware provides a hook for service-wide middleware.
    Middleware(http.Handler) http.Handler
}
```

The 5 service types that are accepted and hostable on the `SimpleServer`:

```go
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
```

Where `JSONEndpoint`, `JSONContextEndpoint`, `ContextHandler` and `ContextHandlerFunc` are defined as:

```go
type JSONEndpoint func(*http.Request) (int, interface{}, error)

type JSONContextEndpoint func(context.Context, *http.Request) (int, interface{}, error)

type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)
```

Also, the one service type that works with an `RPCServer`:

```go
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
```

The `Middleware(..)` functions offer each service a 'hook' to wrap each of its endpoints. This may be handy for adding additional headers or context to the request. This is also the point where other, third-party middleware could be easily plugged in (i.e. oauth, tracing, metrics, logging, etc.)

## The `pubsub` packages

The base `pubsub` package contains three generic interfaces for publishing data to queues and subscribing and consuming data from those queues.

```go
// Publisher is a generic interface to encapsulate how we want our publishers
// to behave. Until we find reason to change, we're forcing all publishers
// to emit protobufs.
type Publisher interface {
    // Publish will publish a message.
    Publish(ctx context.Context, key string, msg proto.Message) error
    // Publish will publish a []byte message.
    PublishRaw(ctx context.Context, key string, msg []byte) error
}

// MultiPublisher is an interface for publishers who support sending multiple
// messages in a single request, in addition to individual messages.
type MultiPublisher interface {
    Publisher

	// PublishMulti will publish multiple messages with a context.
	PublishMulti(context.Context, []string, []proto.Message) error
	// PublishMultiRaw will publish multiple raw byte array messages with a context.
	PublishMultiRaw(context.Context, []string, [][]byte) error
}

// Subscriber is a generic interface to encapsulate how we want our subscribers
// to behave. For now the system will auto stop if it encounters any errors. If
// a user encounters a closed channel, they should check the Err() method to see
// what happened.
type Subscriber interface {
    // Start will return a channel of raw messages.
    Start() <-chan SubscriberMessage
    // Err will contain any errors returned from the consumer connection.
    Err() error
    // Stop will initiate a graceful shutdown of the subscriber connection.
    Stop() error
}
```

Where a `SubscriberMessage` is an interface that gives implementations a hook for acknowledging/delete messages. Take a look at the docs for each implementation in `pubsub` to see how they behave.

There are currently major 4 implementations of the `pubsub` interfaces:

For pubsub via Amazon's SNS/SQS, you can use the `pubsub/aws` package.

For pubsub via Google's Pubsub, you can use the `pubsub/gcp` package. This package offers 2 ways of publishing to Google PubSub. `gcp.NewPublisher` uses the RPC client and `gcp.NewHTTPPublisher` will publish over plain HTTP, which is useful for the App Engine standard environment.

For pubsub via Kafka topics, you can use the `pubsub/kafka` package.

For publishing via HTTP, you can use the `pubsub/http` package.

The `MultiPublisher` interface is only implemented by `pubsub/gcp`.

## The `pubsub/pubsubtest` package

This package contains 'test' implementations of the `pubsub.Publisher`, `pubsub.MultiPublisher`, and `pubsub.Subscriber` interfaces that will allow developers to easily mock out and test their `pubsub` implementations:

```go
type TestPublisher struct {
    // Published will contain a list of all messages that have been published.
    Published []TestPublishMsg

    // GivenError will be returned by the TestPublisher on publish.
    // Good for testing error scenarios.
    GivenError error

    // FoundError will contain any errors encountered while marshalling
    // the protobuf struct.
    FoundError error
}

type TestSubscriber struct {
    // ProtoMessages will be marshalled into []byte  used to mock out
    // a feed if it is populated.
    ProtoMessages []proto.Message

    // JSONMEssages will be marshalled into []byte and used to mock out
    // a feed if it is populated.
    JSONMessages []interface{}

    // GivenErrError will be returned by the TestSubscriber on Err().
    // Good for testing error scenarios.
    GivenErrError error

    // GivenStopError will be returned by the TestSubscriber on Stop().
    // Good for testing error scenarios.
    GivenStopError error

    // FoundError will contain any errors encountered while marshalling
    // the JSON and protobuf struct.
    FoundError error
}
```

## The `web` package

This package contains a handful of very useful functions for parsing types from request queries and payloads.

## Examples

* Several reference implementations utilizing `server` and `pubsub` are available in the ['examples'](https://github.com/NYTimes/gizmo/tree/master/examples) subdirectory.

<sub>The Gizmo logo was based on the Go mascot designed by Renée French and copyrighted under the Creative Commons Attribution 3.0 license.</sub>
