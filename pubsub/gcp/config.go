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
