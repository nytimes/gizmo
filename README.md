# Gizmo Microservice Toolkit [![GoDoc](https://godoc.org/github.com/gizmo/gizmo?status.svg)](https://godoc.org/github.com/nytimes/gizmo) [![Build Status](https://travis-ci.org/NYTimes/gizmo.svg?branch=master)](https://travis-ci.org/NYTimes/gizmo)

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

## The `config` package

This package contains a handful of structs meant for managing common configuration options and credentials. There are currently configs for:

* MySQL
* MongoDB
* Oracle
* AWS (SNS, SQS, S3)
* Kafka
* Gorilla's `securecookie`
* Gizmo Servers

The package also has a generic `Config` type that contains all of the above types. It's meant to be a 'catch all' struct that most applications should be able to use.

`config` also contains functions to load these config structs from JSON files, JSON blobs in Consul k/v or environment variables.

## The `server` package

This package is the bulk of the toolkit and relies on `config` for any managing `Server` implementations. A server must implement the following interface:

```go
// Server is the basic interface that defines what expect from any server.
type Server interface {
    Register(Service) error
    Start() error
    Stop() error
}
```

The package offers 2 server implementations:

`SimpleServer`, which is capable of handling basic HTTP and JSON requests via 3 of the available `Service` implementations: `SimpleService`, `JSONService`, and `MixedService`. A service and these implementations will be defined below.

`RPCServer`, which is capable of serving a gRPC server on one port and JSON endpoints on another. This kind of server can only handle the `RPCService` implementation.

The `Service` interface is minimal to allow for maximum flexibility:
```go
type Service interface {
    Prefix() string

    // Middleware provides a hook for service-wide middleware.
    Middleware(http.Handler) http.Handler
}
```

The 3 service types that are accepted and hostable on the `SimpleServer`:

```go
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
```

Where a `JSONEndpoint` is defined as:

```
type JSONEndpoint func(*http.Request) (int, interface{}, error)
```

Also, the one service type that works with an `RPCServer`:

```go
type RPCService interface {
    Service

    Service() (grpc.ServiceDesc, interface{})

    // route - method - func
    JSONEndpoints() map[string]map[string]JSONEndpoint
    // JSONMiddleware provides a hook for service-wide middleware around JSONEndpoints.
    JSONMiddlware(JSONEndpoint) JSONEndpoint
}
```

The `Middleware(..)` functions offer each service a 'hook' to wrap each of it's endpoints. This may be handy for adding additional headers or context to the request. This is also the point where other, third-party middleware could be easily be plugged in (ie. oauth, tracing, metrics, logging, etc.)

## The `pubsub` package

This package contains two generic interfaces for publishing data to queues and subscribing and consuming data from those queues.

```go
// Publisher is a generic interface to encapsulate how we want our publishers
// to behave. Until we find reason to change, we're forcing all publishers
// to emit protobufs.
type Publisher interface {
    // Publish will publish a message.
    Publish(string, proto.Message) error
    // Publish will publish a []byte message.
    PublishRaw(string, []byte) error
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

Where a `SubscriberMessage` is an interface that gives implementations a hook for acknowledging/delete messages. Take a look at the docs for each implmentation in `pubsub` to see how they behave.

There are currently 2 implementations of each type of `pubsub` interfaces:

For pubsub via Amazon's SNS/SQS, you can use the `SNSPublisher` and the `SQSSubscriber`.

For pubsub via Kafka topics, you can use the `KakfaPublisher` and the `KafkaSubscriber`.

## The `pubsub/pubsubtest` package

This package contains 'test' implementations of the `pubsub.Publisher` and `pubsub.Subscriber` interfaces that will allow developers to easily mock out and test their `pubsub` implementations:

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

* Several reference implementations utilizing `server` and `pubsub` are available in the ['examples'](https://github.com/nytimes/gizmo/tree/master/examples) subdirectory. 

<sub>The Gizmo logo was based on the Go mascot designed by Renée French and copyrighted under the Creative Commons Attribution 3.0 license.</sub>
