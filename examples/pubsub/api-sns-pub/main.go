package main

import (
	"github.com/nytimes/gizmo/examples/pubsub/api-sns-pub/service"

	"github.com/nytimes/gizmo/config/combined"
	"github.com/nytimes/gizmo/server"
)

func main() {
	// showing 1 way of managing gizmo/config: importing from a local file
	cfg := combined.NewConfig("./config.json")

	server.Init("nyt-json-pub-proxy", cfg.Server)

	err := server.Register(service.NewJSONPubService(cfg))
	if err != nil {
		server.Log.Fatal("unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("server encountered a fatal error: ", err)
	}
}
