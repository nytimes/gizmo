package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/examples/pubsub/api-sns-pub/service"
	"github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/NYTimes/gizmo/server"
)

func main() {
	var cfg struct {
		Server *server.Config
		SNS    aws.SNSConfig
	}
	config.LoadJSONFile("./config.json", &cfg)

	server.Init("nyt-json-pub-proxy", cfg.Server)

	err := server.Register(service.NewJSONPubService(cfg.SNS))
	if err != nil {
		server.Log.Fatal("unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("server encountered a fatal error: ", err)
	}
}
