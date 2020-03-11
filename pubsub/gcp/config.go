package gcp

import (
	gpubsub "cloud.google.com/go/pubsub"
	"github.com/kelseyhightower/envconfig"
)

// Config holds common credentials and config values for
// working with GCP PubSub.
type Config struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT"`

	// For publishing
	Topic string `envconfig:"GCP_PUBSUB_TOPIC"`

	// Batch settings for GCP publisher
	// See: https://godoc.org/cloud.google.com/go/pubsub#PublishSettings
	// Notes:
	// This config will not allow you to set zero values for PublishSettings.
	// Applications using these settings should be aware that Publish requests
	// will block until the lowest of the thresholds in PublishSettings is met.
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
