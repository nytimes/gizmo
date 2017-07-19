# The 'Reading List' Example

This example implements a clone of NYT's 'saved articles API' that allows users to save, delete and retrieve nytimes.com article URLs.

This service utilizes Google Cloud Datastore and is set up to be built and published to Google Container Registry, deployed to Google Container Engine and monitored by Google Cloud Tracing.

Instead of utilizing NYT's auth, this example leans on Google OAuth and Google Cloud Endpoints for user identity.

To run locally, have the latest version of `gcloud` installed and execute the `./run_local.sh` script to start up the Datastore emulater and the reading list server.

A few highlights of this service worth calling out:

* [service.yaml](service.yaml)
  * An Open API specification that describes the endpoints in this service.
* [gen-proto.sh](gen-proto.sh)
  * A script that relies on github.com/NYTimes/openapi2proto to generate a gRPC service spec with HTTP annotations from the Open API spec along with the Go/Cloud Endpoint stubs via protoc.
* [service.go](service.go)
  * The actual [kit.Service](http://godoc.org/github.com/NYTimes/gizmo/server/kit#Service) implementation.
* [http_client.go](http_client.go)
  * A go-kit client for programmatically accessing the API via HTTP/JSON.
* [cmd/cli/main.go](cmd/cli/main.go)
  * A CLI wrapper around the gRPC client.
* [.drone.yaml](.drone.yaml)
  * An example configuration file for [Drone CI](http://readme.drone.io/) using the [NYTimes/drone-gke](https://github.com/nytimes/drone-gke) plugin for managing automated deployments to Google Container Engine.
* [cloud-endpoints/service-ce-prd.yaml](cloud-endpoints/service-ce-prd.yaml)
  * A service configuration for Google Cloud Endpoints. 

This example [mirrors an example](https://github.com/NYTimes/marvin/tree/master/examples/reading-list#the-reading-list-example) in gizmo's sibling server for Google App Engine, marvin.
