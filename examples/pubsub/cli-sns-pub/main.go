package main

import (
	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := aws.LoadSNSConfigFromEnv()

	pub, err := aws.NewPublisher(cfg)
	if err != nil {
		pubsub.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("unable to init publisher")
	}

	catArticle := &nyt.SemanticConceptArticle{
		Title:  "It's a Cat World",
		Byline: "By JP Robinson",
		Url:    "http://www.nytimes.com/2015/11/25/its-a-cat-world",
	}

	err = pub.Publish(nil, catArticle.Url, catArticle)
	if err != nil {
		pubsub.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("unable to publish message")
	}

	pubsub.Log.WithFields(logrus.Fields{
		"articles": catArticle,
	}).Info("successfully published cat article")
}
