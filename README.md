# Gizmo Microservice Toolkit [![GoDoc](https://godoc.org/github.com/gizmo/gizmo?status.svg)](https://godoc.org/github.com/NYTimes/gizmo) [![Build Status](https://travis-ci.org/NYTimes/gizmo.svg?branch=master)](https://travis-ci.org/NYTimes/gizmo) [![Coverage Status](https://coveralls.io/repos/NYTimes/gizmo/badge.svg?branch=master&service=github)](https://coveralls.io/github/NYTimes/gizmo?branch=master)

<p align="center">
  <img src="https://lh3.googleusercontent.com/GZ8BtUP_lQJTibMTajjWeLogeXjT5ZzhT_aNRF4WSxNNKQcDDZ3a4k9HfZpXEnfx4T0-wfkv_qjA303Uz-BcJQWpQMHW0U3B-U7H8CI7dDrnxXm0LpCtRN1mrVUSgGAKRe4nSoZxxk-kUpKS_RefWesArYwVEJ62CPucDU5LlKVGUF8XihuWg5aRYQA_5Rwr16tJP5CZECaY0qG-5r6CmRy9gTYbczWmOiFisaQ2QJchYdTEQyrHh4ZH4zfoudOlSisSLQuZyRpVsI5i3uTsS1POL3cMvFoiE2vypv4p5m-G7ZCnFpsKVd4V5FSHE6iBhaJrFm6sF2MyJWBftQxuw4PzOB-wHXhTWwQHK5SWwPCxzgtMCvoC0ECUhx3aCa2Ga2t8Skf4linasSdkl98ttg_TDJTrkBx-W51FEP8CPH_0OZ0BNUOC6v3miAJcAw7eExtyYTK0-bYPOCW-SQ0ooFa_oOBUrddFNpvqh0wzbhGthXTfHnA9ILAm5g5bBAw0ZS8BVZSVLdRFQLgT7dCybNHxblAG9HSW5E14Jed-JvF6D1gxGqiZZnUQ0nByrPbRsoi-FFnr2fSFkeqRaR6JfXeMuKlJ59As6Owq3Q=w278-h306-no"/>
</p>

This toolkit provides packages to put together server and pubsub daemons with the following features:

* Standardized configuration and logging
* Health check endpoints with configurable strategies
* Configuration for managing pprof endpoints and log levels
* Structured logging containing basic request information
* Useful metrics for endpoints
* Graceful shutdowns
* Basic interfaces to define expectations and vocabulary

### Packages

#### [`server`](https://godoc.org/github.com/NYTimes/gizmo/server)

This package is the bulk of the toolkit and relies on `server.Config` to manage `Server` implementations.

It offers 2 server implementations:

1. [`SimpleServer`](https://godoc.org/github.com/NYTimes/gizmo/server#SimpleServer), which is capable of handling basic HTTP and JSON requests via 5 of the available `Service` implementations: `SimpleService`, `JSONService`, `ContextService`, `MixedService` and a `MixedContextService`.

2. [`RPCServer`](https://godoc.org/github.com/NYTimes/gizmo/server#RPCServer), which is capable of serving a gRPC server on one port and JSON endpoints on another. This kind of server can only handle the `RPCService` implementation.

#### [`config`](https://godoc.org/github.com/NYTimes/gizmo/config)

The `config` package contains a handful of useful functions to load to configuration structs from JSON files, JSON blobs in Consul k/v, or environment variables.

* There are also many structs for common configuration options and credentials of different Cloud Services and Databases.
* The package also has a generic `Config` type in the `config/combined` subpackage that contains all of the above types. It's meant to be a 'catch all' convenience struct that many applications should be able to use.

#### [`pubsub`](https://godoc.org/github.com/NYTimes/gizmo/pubsub)

The `pubsub` package contains two (`publisher` and `subscriber`) generic interfaces for publishing data to queues as well as subscribing and consuming data from those queues.

#### [`pubsub/pubsubtest`](https://godoc.org/github.com/NYTimes/gizmo/pubsub/pubsubtest)

This package contains test implementations of the `pubsub.Publisher`, `pubsub.MultiPublisher`, and `pubsub.Subscriber` interfaces that will allow developers to easily mock out and test their `pubsub` implementations.

#### [`web`](https://godoc.org/github.com/NYTimes/gizmo/web)

This package contains a handful of very useful functions for parsing types from request queries and payloads.

#### Examples

* Several reference implementations utilizing `server` and `pubsub` are available in the [`examples`](https://github.com/NYTimes/gizmo/tree/master/examples) subdirectory.
* There are also examples within the GoDoc: [here](https://godoc.org/github.com/NYTimes/gizmo/examples)


<sub>The Gizmo logo was based on the Go mascot designed by Renée French and copyrighted under the Creative Commons Attribution 3.0 license.</sub>
