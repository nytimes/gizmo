package main

import "github.com/NYTimes/gizmo/examples/pubsub/sqs-sub/service"

func main() {
	service.Init()

	service.Log.Info("starting cats subscriber process")

	if err := service.Run(); err != nil {
		service.Log.Fatal("cats subscriber encountered a fatal error: ", err)
	}

	service.Log.Info("cats subscriber process shutting down")
}
