package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/NYTimes/logrotate"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

var (
	Log = logrus.New()

	sub pubsub.Subscriber

	client nyt.Client

	articles []nyt.SemanticConceptArticle
)

type Config struct {
	MostPopularToken string
	SemanticToken    string
	Log              *string
	SQS              aws.SQSConfig
}

func Init() {
	var cfg *Config
	config.LoadJSONFile("./config.json", &cfg)
	config.SetLogOverride(cfg.Log)

	if *cfg.Log != "" {
		lf, err := logrotate.NewFile(*cfg.Log)
		if err != nil {
			Log.Fatalf("unable to access log file: %s", err)
		}
		Log.Out = lf
		Log.Formatter = &logrus.JSONFormatter{}
	} else {
		Log.Out = os.Stderr
	}

	pubsub.Log = Log

	var err error
	client = nyt.NewClient(cfg.MostPopularToken, cfg.SemanticToken)

	sub, err = aws.NewSubscriber(cfg.SQS)
	if err != nil {
		Log.Fatal("unable to init SQS: ", err)
	}
}

func Run() (err error) {
	stream := sub.Start()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		Log.Infof("received kill signal %s", <-ch)
		err = sub.Stop()
	}()

	var article nyt.SemanticConceptArticle
	for msg := range stream {

		if err = proto.Unmarshal(msg.Message(), &article); err != nil {
			Log.Error("unable to unmarshal article from SQS: ", err)
			if err = msg.Done(); err != nil {
				Log.Error("unable to delete message from SQS: ", err)
			}
			continue
		}

		// do something!
		fmt.Println("Most Recent Article on 'Cats':")
		out, _ := json.MarshalIndent(article, "", "    ")
		fmt.Fprint(os.Stdout, string(out))
		articles = append(articles, article)

		if err = msg.Done(); err != nil {
			Log.WithFields(logrus.Fields{
				"article": article,
			}).Error("unable to delete message from SQS: ", err)
		}
	}

	return err
}

func metricsNamespace() string {
	// get only server base name
	name, _ := os.Hostname()
	name = strings.SplitN(name, ".", 2)[0]
	// set it up to be paperboy.servername
	name = strings.Replace(name, "-", ".", 1)
	// add the 'apps' prefix  to keep things neat
	return "apps." + name + ".cats-subscriber"
}
