package gcp

import (
	"github.com/kelseyhightower/envconfig"
	gpubsub "cloud.google.com/go/pubsub"
)

// Config holds common credentials and config values for
// working with GCP PubSub.
type Config struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT"`

	// For publishing
	Topic string `envconfig:"GCP_PUBSUB_TOPIC"`
	
	// Batch settings for GCP publisher		   
	// See: https://godoc.org/cloud.google.com/go/pubsub#PublishSettings
	// Note: this config will not allow you to go lower than the 
	// default PublishSettings values
	PublishSettings gpubsub.PublishSettings

	// For subscribing
    Subscription string `envconfig:"GCP_PUBSUB_SUBSCRIPTION"`
}

// LoadConfigFromEnv will attempt to load a PubSub config
// from environment variables.
func LoadConfigFromEnv() Config {
	var ps Config
	envconfig.Process("", &ps)
	return ps
}
