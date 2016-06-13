package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/kelseyhightower/envconfig"
)

// EnvAppName is used as a prefix for environment variable
// names when using the LoadXFromEnv funcs.
// It defaults to empty.
var EnvAppName = ""

// LoadEnvConfig will use envconfig to load the
// given config struct from the environment.
func LoadEnvConfig(c interface{}) {
	err := envconfig.Process(EnvAppName, c)
	if err != nil {
		log.Fatalf("unable to load env variable for %T: %s", c, err)
	}
}

// LoadJSONFile is a helper function to read a config file into whatever
// config struct you need. For example, your custom config could be composed
// of one or more of the given Config, AWS, MySQL, Oracle or MongoDB structs.
func LoadJSONFile(fileName string, cfg interface{}) {
	cb, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read config file '%s': %s", fileName, err)
	}

	if err = json.Unmarshal(cb, &cfg); err != nil {
		log.Fatalf("Unable to parse JSON in config file '%s': %s", fileName, err)
	}
}

// LoadJSONFromConsulKV is a helper function to read a JSON string found
// in a path defined by configKey inside Consul's Key Value storage then
// unmarshalled into a config struct, like LoadJSONFile does.
// It assumes that the Consul agent is running with the default setup,
// where the HTTP API is found via 127.0.0.1:8500.
func LoadJSONFromConsulKV(configKeyParameter string, cfg interface{}) interface{} {
	configKeyParameterValue := strings.SplitN(configKeyParameter, ":", 2)
	if len(configKeyParameterValue) < 2 {
		log.Fatalf("Undefined Consul KV configuration path. It should be defined using the format consul:path/to/JSON/string")
	}
	configKey := configKeyParameterValue[1]
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Unable to setup Consul client: %s", err)
	}
	kv := client.KV()
	kvPair, _, err := kv.Get(configKey, nil)
	if err != nil {
		log.Fatalf("Unable to read config in key '%s' from Consul KV: %s", configKey, err)
	}
	if kvPair == nil {
		log.Fatalf("Undefined key '%s' in Consul KV", configKey)
	}
	if len(kvPair.Value) == 0 {
		log.Fatalf("Empty JSON in Consul KV for key '%s'", configKey)
	}
	if err = json.Unmarshal(kvPair.Value, &cfg); err != nil {
		log.Fatalf("Unable to parse JSON in Consul KV for key '%s': %s", configKey, err)
	}
	return cfg
}
