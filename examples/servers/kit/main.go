package main

import (
	"github.com/nytimes/gizmo/config"
	"github.com/nytimes/gizmo/server/kit"

	"github.com/nytimes/gizmo/examples/servers/kit/api"
)

func main() {
	var cfg api.Config
	config.LoadEnvConfig(&cfg)

	// runs the HTTP _AND_ gRPC servers
	err := kit.Run(api.New(cfg))
	if err != nil {
		panic("problems running service: " + err.Error())
	}
}
