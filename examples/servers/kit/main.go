package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server/kit"

	"github.com/NYTimes/gizmo/examples/servers/kit/api"
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
