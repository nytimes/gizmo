## `mysql-saved-items`
* example using a `gizmo/server.JSONService` with a `gizmo/server.SimpleServer` using MySQL for persistence.
* this is a simple implementation of [nytimes.com/saveditems](http://www.nytimes.com/saveditems). It provides 3 endpoints to create, delete and list 'saved items' for a single user.

### The config in this example is loaded via a local JSON file and a custom config struct that is composed of a `gizmo/config.Config` struct.
