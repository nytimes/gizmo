## `rpc`
* example using the experimental `gizmo/server.RPCService` and `gizmo/server.RPCServer`.
* one endpoint will serve the most popular NYT articles and another will serve a listing recent articles in The New York Times about 'cats'.
* this service will expose both the 'most popular' and 'cats' endpoints via gRPC _and_ simple JSON using two separate ports.

### The config in this example is loaded via a local JSON file and a custom config struct that is composed of a `gizmo/config.Server` struct.
