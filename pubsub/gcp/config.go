package gcp

import (
	"github.com/nytimes/gizmo/config"
	"github.com/nytimes/gizmo/config/gcp"
)

// Config holds common credentials and config values for
// working with GCP PubSub.
type Config struct {
	gcp.Config

	// For publishing
	Topic string `envconfig:"GCP_PUBSUB_TOPIC"`
	// For subscribing
	Subscription string `envconfig:"GCP_PUBSUB_SUBSCRIPTION"`
}

// LoadConfigFromEnv will attempt to load a PubSub config
// from environment variables.
func LoadConfigFromEnv() Config {
	var ps Config
	config.LoadEnvConfig(&ps)
	return ps
}
