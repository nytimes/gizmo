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

It offers 1 server implementation:

[`SimpleServer`](https://godoc.org/github.com/NYTimes/gizmo/server#SimpleServer), which is capable of handling basic HTTP and JSON requests via 5 of the available `Service` implementations: `SimpleService`, `JSONService`, `ContextService`, `MixedService` and a `MixedContextService`.

#### [`server/kit`](https://godoc.org/github.com/NYTimes/gizmo/server/kit)

The `server/kit` package embodies Gizmo's goals to combine with go-kit.

* In this package you'll find:
    * A more opinionated server with fewer choices.
    * go-kit used for serving HTTP/JSON & gRPC used for serving HTTP2/RPC
    * Monitoring and metrics are automatically registered if running within App Engine.
    * Logs go to stdout locally or directly to Stackdriver when in GCP.
    * Using Go's 1.8 graceful HTTP shutdown.
    * Services using this package are expected to deploy to GCP.

#### [`auth`](https://godoc.org/github.com/NYTimes/gizmo/auth)

The `auth` package provides primitives for verifying inbound authentication tokens:

* The `PublicKeySource` interface is meant to provide `*rsa.PublicKeys` from JSON Web Key Sets.
* The `Verifier` struct composes key source implementations with custom decoders and verifier functions to streamline server side token verification.

#### [`auth/gcp`](https://godoc.org/github.com/NYTimes/gizmo/auth/gcp)

The `auth/gcp` package provides 2 Google Cloud Platform based `auth.PublicKeySource` and `oauth2.TokenSource` implementations:

* The "Identity" key source and token source rely on GCP's [identity JWT mechanism for asserting instance identities](https://cloud.google.com/compute/docs/instances/verifying-instance-identity). This is the preferred method for asserting instance identity on GCP.
* The "IAM" key source and token source rely on GCP's IAM services for [signing](https://cloud.google.com/iam/reference/rest/v1/projects.serviceAccounts/signJwt) and [verifying JWTs](https://cloud.google.com/iam/reference/rest/v1/projects.serviceAccounts.keys/get). This method can be used outside of GCP, if needed and can provide a bridge for users transitioning from the 1st generation App Engine (where Identity tokens are not available) runtime to the 2nd.


#### [`config`](https://godoc.org/github.com/NYTimes/gizmo/config)

The `config` package contains a handful of useful functions to load to configuration structs from JSON files or environment variables.

There are also many structs for common configuration options and credentials of different Cloud Services and Databases.

#### [`pubsub`](https://godoc.org/github.com/NYTimes/gizmo/pubsub)
 
The `pubsub` package contains two (`publisher` and `subscriber`) generic interfaces for publishing data to queues as well as subscribing and consuming data from those queues.

There are 4 implementations of `pubsub` interfaces:

* For pubsub via Amazon's SNS/SQS, you can use the [`pubsub/aws`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/aws) package

* For pubsub via Google's Pubsub, you can use the [`pubsub/gcp`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/gcp) package

* For pubsub via Kafka topics, you can use the [`pubsub/kafka`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/kafka) package

* For publishing via HTTP, you can use the [`pubsub/http`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/http) package


#### [`pubsub/pubsubtest`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/pubsubtest)

The `pubsub/pubsubtest` package contains test implementations of the `pubsub.Publisher`, `pubsub.MultiPublisher`, and `pubsub.Subscriber` interfaces that will allow developers to easily mock out and test their `pubsub` implementations.

#### Examples

* Several reference implementations utilizing `server` and `pubsub` are available in the [`examples`](https://github.com/NYTimes/gizmo/tree/master/examples) subdirectory.
* There are also examples within the GoDoc: [here](https://godoc.org/github.com/NYTimes/gizmo/examples)

<sub><strong>If you experience any issues please create an issue and/or reach out on the #gizmo channel in the [Gophers Slack Workspace](https://invite.slack.golangbridge.org) with what you've found.</strong></sub>

<sub>The Gizmo logo was based on the Go mascot designed by Renée French and copyrighted under the Creative Commons Attribution 3.0 license.</sub>
