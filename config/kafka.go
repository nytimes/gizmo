package config

import "strings"

// Kafka holds the basic information for working with Kafka
type Kafka struct {
	BrokerHosts []string
	// BrokerHostsString is used when loading the list from environment variables.
	// If loaded via the LoadEnvConfig() func, BrokerHosts will get updated with these
	// values.
	BrokerHostsString string `envconfig:"KAFKA_BROKER_HOSTS"`

	Partition int32  `envconfig:"KAFKA_PARTITION"`
	Topic     string `envconfig:"KAFKA_TOPIC"`

	MaxRetry int `envconfig:"KAFKA_MAX_RETRY"`
}

// LoadKafkaFromEnv will attempt to load an Kafka object
// from environment variables. If not populated, nil
// is returned.
func LoadKafkaFromEnv() *Kafka {
	var kafka Kafka
	LoadEnvConfig(&kafka)
	if kafka.BrokerHostsString == "" {
		return nil
	}
	kafka.BrokerHosts = strings.Split(kafka.BrokerHostsString, ",")
	return &kafka
}
