## `simple`
* example using a `gizmo/server.SimpleService` with a `gizmo/server.SimpleServer`.
* one endpoint will serve JSON of the most popular NYT articles and another will serve an HTML page listing recent articles in The New York Times about 'cats'.

### The config in this example is loaded via environment variables and a custom config struct that is composed of a `gizmo/config.Server` struct. To load the config, `source` the local `.env` file before running.
