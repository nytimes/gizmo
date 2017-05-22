package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/examples/servers/kit/httpsvc"
	"github.com/NYTimes/gizmo/server/kithttp"
)

func main() {
	var cfg httpsvc.Config
	config.LoadEnvConfig(&cfg)

	err := kithttp.Run(httpsvc.New(cfg))
	if err != nil {
		panic("problems running service: " + err.Error())
	}
}
