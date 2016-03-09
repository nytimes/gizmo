package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/kelseyhightower/envconfig"
)

type (
	// Config is a generic struct to hold information for applications that
	// need to connect to databases, handle cookies, log events or emit metrics.
	// If you have a use case that does not fit this struct, you can
	// make a struct containing just the types that suit your needs and use
	// some of the helper functions in this package to load it from the environment.
	Config struct {
		Server *Server

		AWS         *AWS
		SQS         *SQS
		SNS         *SNS
		S3          *S3
		DynamoDB    *DynamoDB
		ElastiCache *ElastiCache

		Kafka *Kafka

		Oracle *Oracle

		MySQL      *MySQL
		MySQLSlave *MySQL

		MongoDB *MongoDB

		Cookie *Cookie

		GraphiteHost *string `envconfig:"GRAPHITE_HOST"`

		LogLevel *string `envconfig:"APP_LOG_LEVEL"`
		Log      *string `envconfig:"APP_LOG"`
	}

	// Cookie holds information for creating
	// a secure cookie.
	Cookie struct {
		HashKey  string `envconfig:"COOKIE_HASH_KEY"`
		BlockKey string `envconfig:"COOKIE_BLOCK_KEY"`
		Domain   string `envconfig:"COOKIE_DOMAIN"`
		Name     string `envconfig:"COOKIE_NAME"`
	}
)

// EnvAppName is used as a prefix for environment variable
// names when using the LoadXFromEnv funcs.
// It defaults to empty.
var EnvAppName = ""

// LoadConfigFromEnv will attempt to inspect the environment
// of any valid config options and will return a populated
// Config struct with what it found.
// If you need a unique config object and want to use envconfig, you
// will need to run the LoadXXFromEnv for each child struct in
// your config struct. For an example on how to do this, check out the
// guts of this function.
func LoadConfigFromEnv() *Config {
	var app Config
	LoadEnvConfig(&app)
	app.AWS, app.SNS, app.SQS, app.S3, app.DynamoDB, app.ElastiCache = LoadAWSFromEnv()
	app.MongoDB = LoadMongoDBFromEnv()
	app.Kafka = LoadKafkaFromEnv()
	app.MySQL = LoadMySQLFromEnv()
	app.Oracle = LoadOracleFromEnv()
	app.Cookie = LoadCookieFromEnv()
	app.Server = LoadServerFromEnv()
	return &app
}

// LoadEnvConfig will use envconfig to load the
// given config struct from the environment.
func LoadEnvConfig(c interface{}) {
	err := envconfig.Process(EnvAppName, c)
	if err != nil {
		log.Fatal("unable to load env variable: ", err)
	}
}

// LoadCookieFromEnv will attempt to load an Cookie object
// from environment variables. If not populated, nil
// is returned.
func LoadCookieFromEnv() *Cookie {
	var cookie Cookie
	LoadEnvConfig(&cookie)
	if cookie.Name != "" {
		return &cookie
	}
	return nil
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

// NewConfig will attempt to unmarshal the contents
// of the given JSON string source into a Config struct.
// The value of fileName can be either the path to a JSON
// file or a path to a JSON string found in Consul's Key
// Value storage (using the format consul:path/to/JSON/string).
// If the value of fileName is empty, a blank Config will
// be returned.
func NewConfig(fileName string) *Config {
	var c Config
	if fileName == "" {
		return &c
	}
	if strings.HasPrefix(fileName, "consul") {
		return LoadJSONFromConsulKV(fileName, &c).(*Config)
	}
	LoadJSONFile(fileName, &c)
	return &c
}
