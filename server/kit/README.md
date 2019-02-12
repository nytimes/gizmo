# Welcome to the 2nd generation of Gizmo servers 🚀!

Gizmo's intentions from the beginning were to eventually join forces with the wonders of the [go-kit toolkit](https://github.com/go-kit/kit). This package is meant to embody that goal.

The `kit` server is composed of multiple [kit/transport/http.Servers](https://godoc.org/github.com/go-kit/kit/transport/http#Server) that are tied together with a common HTTP mux, HTTP options and middlewares. By default all HTTP endpoints will be encoded as JSON, but developers may override each [HTTPEndpoint](https://godoc.org/github.com/NYTimes/gizmo/server/kit#HTTPEndpoint) to use whatever encoding they need. If users need to use gRPC, they can can register the same endpoints to serve both HTTP and gRPC requests on two different ports.

This server expects to be configured via environment variables. The available variables can be found by inspecting the [Config struct](https://godoc.org/github.com/NYTimes/gizmo/server/kit#Config) within this package. If no health check or [warm up](https://cloud.google.com/appengine/docs/standard/go111/how-instances-are-managed#warmup_requests) endpoints are defined, this server will automatically register basic endpoints there to return a simple "200 OK" response.

Since NYT uses Google Cloud, deploying this server to that environment provides additional perks:

* If running in the [App Engine 2nd Generation runtime (Go >=1.11)](https://cloud.google.com/appengine/docs/standard/go111/), servers will:
  * Automatically catch any panics and send them to Stackdriver Error reporting
  * Automatically use Stackdriver logging and, if `kit.LogXXX` functions are used, logs will be trace enabled and will be combined with their parent access log in the Stackdriver logging console.
  * Automatically register Stackdriver exporters for Open Census trace and monitoring. Most Google Cloud clients (like [Cloud Spanner](https://godoc.org/cloud.google.com/go/spanner)) will detect this and emit the traces. Users can also add their own trace and monitoring spans via [the Open Census clients](https://godoc.org/go.opencensus.io/trace#example-StartSpan).
  * Monitoring, traces and metrics are automatically registered if running within App Engine, Kubernetes Engine, Compute Engine or AWS EC2 Instances. To change the name and version for Error reporting and Traces use `SERVICE_NAME` and `SERVICE_VERSION` environment variables.


For an example of how to build a server that utilizes this package, see the [Reading List example](https://github.com/NYTimes/gizmo/tree/master/examples/servers/reading-list#the-reading-list-example).
