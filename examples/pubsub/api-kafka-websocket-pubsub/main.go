package main

import (
	"github.com/NYTimes/gizmo/config/combined"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/server"

	"github.com/NYTimes/gizmo/examples/pubsub/api-kafka-websocket-pubsub/service"
)

func main() {
	cfg := combined.NewConfig("./config.json")

	// set the pubsub's Log to be the same as server's
	pubsub.Log = server.Log

	// in case we want to override the port or log location via CLI
	server.SetConfigOverrides(cfg.Server)

	server.Init("gamestream-example", cfg.Server)

	err := server.Register(service.NewStreamService(cfg.Server.HTTPPort, cfg.Kafka))
	if err != nil {
		server.Log.Fatal(err)
	}

	if err = server.Run(); err != nil {
		server.Log.Fatal(err)
	}
}
