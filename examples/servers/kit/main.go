package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/examples/servers/kit/api"
	kitserver "github.com/NYTimes/gizmo/server/kit"
)

func main() {
	var cfg api.Config
	config.LoadEnvConfig(&cfg)

	// runs the HTTP _AND_ gRPC servers
	err := kitserver.Run(api.New(cfg))
	if err != nil {
		panic("problems running service: " + err.Error())
	}
}
