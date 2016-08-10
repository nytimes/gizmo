package gcp

import (
	"io/ioutil"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/NYTimes/gizmo/config"
)

// Config holds common Google Cloud Platform credentials.
type Config struct {
	ProjectID string `envconfig:"GCP_PROJECT_ID"`

	// JSONAuthPath points to a file containing a JWT JSON config.
	// This is meant to be a fall back for development environments.
	JSONAuthPath string `envconfig:"GCP_JSON_AUTH_PATH"`

	// Token is a JWT JSON config and may be needed for container
	// environments.
	Token string `envconfig:"GCP_AUTH_TOKEN"`

	// FlexibleVM tells the config we are using a 'flexible' App Engine VM
	// and to use appengine.BackgroundContext()
	FlexibleVM bool `envconfig:"GCP_FLEXIBLE_VM"`
}

// LoadConfigFromEnv will attempt to load a GCP config
// from environment variables.
func LoadConfigFromEnv() Config {
	var gcp Config
	config.LoadEnvConfig(&gcp)
	return gcp
}

// ClientOption will attempt create a new option.ClientOption from
// a the Token or JSONAuthPath fields if provided. If the FlexibleAE flag
// is set to designate this is a 'flexible' App Engine VM,
// just the scope passed in will be used. Otherwise, this function
// assumes you're running on GCE and tacks on a compute.ComputeScope.
func (g Config) ClientOption(scopes ...string) (option.ClientOption, error) {
	if len(g.Token) > 0 {
		return g.optionFromToken(scopes...)
	}

	if len(g.JSONAuthPath) > 0 {
		return g.optionFromJSON(scopes...)
	}

	if g.FlexibleVM {
		return option.WithScopes(scopes...), nil
	}

	if len(scopes) == 0 {
		scopes = append(scopes, compute.ComputeScope)
	}

	return option.WithScopes(scopes...), nil
}

func (g Config) optionFromToken(scopes ...string) (option.ClientOption, error) {
	conf, err := google.JWTConfigFromJSON(
		[]byte(g.Token),
		scopes...,
	)
	if err != nil {
		log.Print("probs with token:", g.Token)
		return nil, err
	}

	return option.WithTokenSource(conf.TokenSource(context.Background())), nil
}

func (g Config) optionFromJSON(scopes ...string) (option.ClientOption, error) {
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
	return option.WithTokenSource(conf.TokenSource(context.Background())), nil
}
