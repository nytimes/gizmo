package main

import (
	"context"
	"fmt"

	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	cfg := gcp.LoadConfigFromEnv()

	pub, err := gcp.NewPublisher(ctx, cfg)
	if err != nil {
		pubsub.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("unable to init publisher")
	}

	keys := []string{
		"https://www.nytimes.com/2018/06/04/its-a-dog-world",
		"https://www.nytimes.com/2018/06/05/big-big-dog-big-big-world",
	}
	messages := []proto.Message{
		&nyt.SemanticConceptArticle{
			Title:  "It's a Dog World",
			Byline: "By David López",
			Url:    "https://www.nytimes.com/2018/06/04/its-a-dog-world",
		},
		&nyt.SemanticConceptArticle{
			Title:  "I'm a big big dog, in a big big world",
			Byline: "By David López",
			Url:    "https://www.nytimes.com/2018/06/05/big-big-dog-big-big-world",
		},
	}
	err = pub.PublishMulti(ctx, keys, messages)
	if err != nil {
		pubsub.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("unable to publish")
	}

	sub, err := gcp.NewSubscriber(ctx, cfg.ProjectID, cfg.Subscription)
	if err != nil {
		pubsub.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("unable to init subscriber")
	}

	pipe := sub.Start()
	for i := 0; i < 2; i++ {
		gotMsg := <-pipe
		fmt.Println("Received message", string(gotMsg.Message()))
		gotMsg.Done()
	}
	sub.Stop()
}
