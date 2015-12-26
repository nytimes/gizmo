package main

import (
	"net/http"
	"time"

	"github.com/NYTimes/gizmo/examples/servers/simple/service"
	"github.com/Sirupsen/logrus"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
)

func main() {
	// showing 1 way of managing gizmo/config: importing from the environment
	cfg := service.Config{Server: &config.Server{}}
	config.LoadEnvConfig(&cfg)
	config.LoadEnvConfig(cfg.Server)

	server.Init("nyt-simple-proxy", cfg.Server)

	server.RegisterMiddleware(timedMiddleware)
	err := server.Register(service.NewSimpleService(&cfg))
	if err != nil {
		server.Log.Fatal("unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("server encountered a fatal error: ", err)
	}
}

func timedMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		h.ServeHTTP(w, r)

		elapsed := time.Since(startTime)
		server.Log.WithFields(logrus.Fields{
			"duration": int64(elapsed / time.Millisecond),
		}).Info("Request completed")
	})
}
