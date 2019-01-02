package main

import (
	"github.com/NYTimes/gizmo/examples/servers/simple/service"
	"github.com/NYTimes/gizmo/server"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	// showing 1 way of managing gizmo/config: importing from the environment
	var cfg service.Config
	envconfig.Process("", &cfg)
	cfg.Server = &server.Config{}
	envconfig.Process("", &cfg.Server)

	server.Init("nyt-simple-proxy", cfg.Server)

	err := server.Register(service.NewSimpleService(&cfg))
	if err != nil {
		server.Log.Fatal("unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("server encountered a fatal error: ", err)
	}
}
