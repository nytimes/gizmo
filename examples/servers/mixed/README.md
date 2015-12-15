## `mixed`
* example using a `gizmo/server.MixedService` with a `gizmo/server.SimpleServer`.
* one endpoint will serve JSON of the most popular NYT articles and another will serve an HTML page listing recent articles in The New York Times about 'cats'. 
* this example is very similar to the `simple` example, but makes use of the available `JSONMiddleware` method in a `gizmo/server.MixedService`.

### The config in this example is loaded via a local JSON file and a custom config struct that is composed of a `gizmo/config.Server` struct.
