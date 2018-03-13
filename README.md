# Gizmo Microservice Toolkit [![GoDoc](https://godoc.org/github.com/gizmo/gizmo?status.svg)](https://godoc.org/github.com/NYTimes/gizmo) [![Build Status](https://travis-ci.org/NYTimes/gizmo.svg?branch=master)](https://travis-ci.org/NYTimes/gizmo) [![Coverage Status](https://coveralls.io/repos/NYTimes/gizmo/badge.svg?branch=master&service=github)](https://coveralls.io/github/NYTimes/gizmo?branch=master)

<p align="center">
  <img src="http://graphics8.nytimes.com/images/blogs/open/2015/gizmo.png"/>
</p>

This toolkit provides packages to put together server and pubsub daemons with the following features:

* Standardized configuration and logging
* Health check endpoints with configurable strategies
* Configuration for managing pprof endpoints and log levels
* Basic interfaces to define expectations and vocabulary
* Structured logging containing basic request information
* Useful metrics for endpoints
* Graceful shutdowns

### Packages

#### [`server`](https://godoc.org/github.com/NYTimes/gizmo/server)

The `server` package is the bulk of the toolkit and relies on `server.Config` to manage `Server` implementations.

It offers 2 server implementations:

1. [`SimpleServer`](https://godoc.org/github.com/NYTimes/gizmo/server#SimpleServer), which is capable of handling basic HTTP and JSON requests via 5 of the available `Service` implementations: `SimpleService`, `JSONService`, `ContextService`, `MixedService` and a `MixedContextService`.

2. [`RPCServer`](https://godoc.org/github.com/NYTimes/gizmo/server#RPCServer), which is capable of serving a gRPC server on one port and JSON endpoints on another. This kind of server can only handle the `RPCService` implementation.


#### [`server/kit`](https://godoc.org/github.com/NYTimes/gizmo/server/kit)

This is an experimental package in Gizmo!

* The rationale behind this package:
    * A more opinionated server with fewer choices
    * go-kit is used for serving HTTP/JSON & gRPC is used for serving HTTP2/RPC
    * Monitoring and metrics are handled by a sidecar (ie. Cloud Endpoints)
    * Logs always go to stdout/stderr
    * Using Go's 1.8 graceful HTTP shutdown
    * Services using this package are meant for deploy to GCP with GKE and Cloud Endpoints.


#### [`config`](https://godoc.org/github.com/NYTimes/gizmo/config)

The `config` package contains a handful of useful functions to load to configuration structs from JSON files, JSON blobs in Consul k/v, or environment variables.

* There are also many structs for common configuration options and credentials of different Cloud Services and Databases.
* The package also has a generic `Config` type in the `config/combined` subpackage that contains all of the above types. It's meant to be a 'catch all' convenience struct that many applications should be able to use.

#### [`pubsub`](https://godoc.org/github.com/NYTimes/gizmo/pubsub)
 
The `pubsub` package contains two (`publisher` and `subscriber`) generic interfaces for publishing data to queues as well as subscribing and consuming data from those queues.

There are 4 implementations of `pubsub` interfaces:

* For pubsub via Amazon's SNS/SQS, you can use the [`pubsub/aws`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/aws) package

* For pubsub via Google's Pubsub, you can use the [`pubsub/gcp`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/gcp) package

* For pubsub via Kafka topics, you can use the [`pubsub/kafka`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/kafka) package

* For publishing via HTTP, you can use the [`pubsub/http`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/http) package



#### [`pubsub/pubsubtest`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/pubsubtest)

The `pubsub/pubsubtest` package contains test implementations of the `pubsub.Publisher`, `pubsub.MultiPublisher`, and `pubsub.Subscriber` interfaces that will allow developers to easily mock out and test their `pubsub` implementations.

#### [`web`](https://godoc.org/github.com/NYTimes/gizmo/web)

The `web` package has a handful of very useful functions for parsing types from request queries and payloads.

#### Examples

* Several reference implementations utilizing `server` and `pubsub` are available in the [`examples`](https://github.com/NYTimes/gizmo/tree/master/examples) subdirectory.
* There are also examples within the GoDoc: [here](https://godoc.org/github.com/NYTimes/gizmo/examples)

<sub><strong>If you experience any issues please create an issue and/or reach out on the #gizmo channel in the [Gophers Slack Workspace](https://invite.slack.golangbridge.org) with what you've found.</strong></sub>

<sub>The Gizmo logo was based on the Go mascot designed by Renée French and copyrighted under the Creative Commons Attribution 3.0 license.</sub>
