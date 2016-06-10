package config

import (
	"io/ioutil"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/cloud"
)

type (

	// GCP holds common Google Cloud Platform credentials.
	GCP struct {
		ProjectID string `envconfig:"GCP_PROJECT_ID" json:"GCP_PROJECT_ID"`

		// JSONAuthPath points to a file containing a JWT JSON config.
		// This is meant to be a fall back for development environments.
		JSONAuthPath string `envconfig:"GCP_JSON_AUTH_PATH" json:"GCP_JSON_AUTH_PATH"`

		// Token is a JWT JSON config and may be needed for container
		// environments.
		Token string `envconfig:"GCP_AUTH_TOKEN" json:"GCP_AUTH_TOKEN"`
	}

	// PubSub holds common credentials and config values for
	// working with GCP PubSub.
	PubSub struct {
		GCP

		// For publishing
		Topic string `envconfig:"GCP_PUBSUB_TOPIC" json:"GCP_PUBSUB_TOPIC"`
		// For subscribing
		Subscription string `envconfig:"GCP_PUBSUB_SUBSCRIPTION" json:"GCP_PUBSUB_SUBSCRIPTION"`
	}
)

// LoadGCPFromEnv will attempt to load a GCP config
// from environment variables.
func LoadGCPFromEnv() GCP {
	var gcp GCP
	LoadEnvConfig(&gcp)
	return gcp
}

// LoadPubSubFromEnv will attempt to load a PubSub config
// from environment variables.
func LoadPubSubFromEnv() PubSub {
	var ps PubSub
	LoadEnvConfig(&ps)
	return ps
}

// NewContext will check attempt to create a new context from
// a the Token or JSONAuthPath fields if provided, otherwise
// google.DefaultClient will be used.
func (g GCP) NewContext(scopes ...string) (context.Context, error) {
	if len(g.Token) > 0 {
		return g.contextFromToken(scopes...)
	}

	if len(g.JSONAuthPath) > 0 {
		return g.contextFromJSON(scopes...)
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
