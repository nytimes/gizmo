package config

import (
	"io/ioutil"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
	"google.golang.org/cloud/pubsub"
)

type (
	GCP struct {
		ProjectID string `envconfig:"GCP_PROJECT_ID"`

		JSONAuthPath string `envconfig:"GCP_JSON_AUTH_PATH"`

		Token string `envconfig:"GCP_AUTH_TOKEN"`
	}

	PubSub struct {
		GCP
		Topic        string `envconfig:"GCP_PUBSUB_TOPIC"`
		Subscription string `envconfig:"GCP_PUBSUB_SUBSCRIPTION"`
	}

	Datastore struct {
		GCP
	}
)

// LoadPubSubFromEnv will attempt to load a Metrics object
// from environment variables.
func LoadPubSubFromEnv() PubSub {
	var ps PubSub
	LoadEnvConfig(&ps)
	return ps
}

// LoadDatastoreFromEnv will attempt to load a Metrics object
// from environment variables.
func LoadDatastoreFromEnv() Datastore {
	var ds Datastore
	LoadEnvConfig(&ds)
	return ds
}

func (d Datastore) NewContext() (context.Context, error) {
	return d.GCP.NewContext(datastore.ScopeDatastore)
}

func (d Datastore) NewClient(ctx context.Context) (*datastore.Client, error) {
	return datastore.NewClient(ctx, d.ProjectID)
}

func (p PubSub) NewContext() (context.Context, error) {
	return p.GCP.NewContext(pubsub.ScopePubSub)
}

func (g GCP) NewContext(scopes ...string) (context.Context, error) {
	log.Printf("gcp conf? %#v", g)
	if len(g.JSONAuthPath) > 0 {
		return g.contextFromJSON(scopes...)
	}

	if len(g.Token) > 0 {
		return g.contextFromToken(scopes...)
	}

	if len(scopes) == 0 {
		scopes = append(scopes, compute.ComputeScope)
	}

	client, err := google.DefaultClient(oauth2.NoContext, scopes...)
	if err != nil {
		return nil, err
	}
	return cloud.NewContext(g.ProjectID, client), nil
}

func (g GCP) contextFromToken(scopes ...string) (context.Context, error) {
	conf, err := google.JWTConfigFromJSON(
		[]byte(g.Token),
		scopes...,
	)
	if err != nil {
		log.Print("probs with token:", g.Token)
		return nil, err
	}

	return cloud.NewContext(g.ProjectID, conf.Client(oauth2.NoContext)), nil
}

func (g GCP) contextFromJSON(scopes ...string) (context.Context, error) {
	jsonKey, err := ioutil.ReadFile(g.JSONAuthPath)
	if err != nil {
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		scopes...,
	)
	if err != nil {
		return nil, err
	}

	return cloud.NewContext(g.ProjectID, conf.Client(oauth2.NoContext)), nil
}
