package main

import (
	"github.com/nytimes/gizmo/examples/servers/simple/service"

	"github.com/nytimes/gizmo/config"
	"github.com/nytimes/gizmo/server"
)

func main() {
	// showing 1 way of managing gizmo/config: importing from the environment
	var cfg service.Config
	config.LoadEnvConfig(&cfg)
	cfg.Server = &server.Config{}
	config.LoadEnvConfig(cfg.Server)

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
