package combined

import (
	"strings"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/config/aws"
	"github.com/NYTimes/gizmo/config/cookie"
	"github.com/NYTimes/gizmo/config/metrics"
	"github.com/NYTimes/gizmo/config/mongodb"
	"github.com/NYTimes/gizmo/config/mysql"
	"github.com/NYTimes/gizmo/config/oracle"
	"github.com/NYTimes/gizmo/config/postgresql"
	awsps "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/NYTimes/gizmo/pubsub/kafka"
	"github.com/NYTimes/gizmo/server"
)

// Config is a generic struct to hold information for applications that
// need to connect to databases, handle cookies, log events or emit metrics.
// If you have a use case that does not fit this struct, you can
// make a struct containing just the types that suit your needs and use
// some of the helper functions in this package to load it from the environment.
type Config struct {
	Server *server.Config

	AWS         aws.Config
	SQS         awsps.SQSConfig
	SNS         awsps.SNSConfig
	S3          aws.S3
	DynamoDB    aws.DynamoDB
	ElastiCache aws.ElastiCache

	Kafka *kafka.Config

	Oracle oracle.Config

	PostgreSQL *postgresql.Config

	MySQL      *mysql.Config
	MySQLSlave *mysql.Config

	MongoDB *mongodb.Config

	Cookie cookie.Config

	// GraphiteHost is DEPRECATED. Please use the
	// Metrics config with "Type":"graphite" and this
	// value in the "Addr" field.
	GraphiteHost *string `envconfig:"GRAPHITE_HOST"`

	Metrics metrics.Config

	LogLevel *string `envconfig:"APP_LOG_LEVEL"`
	Log      *string `envconfig:"APP_LOG"`
}

// LoadConfigFromEnv will attempt to inspect the environment
// of any valid config options and will return a populated
// Config struct with what it found.
// If you need a unique config object and want to use envconfig, you
// will need to run the LoadXXFromEnv for each child struct in
// your config struct. For an example on how to do this, check out the
// guts of this function.
func LoadConfigFromEnv() *Config {
	var app Config
	config.LoadEnvConfig(&app)
	app.AWS = aws.LoadConfigFromEnv()
	app.SNS = awsps.LoadSNSConfigFromEnv()
	app.SQS = awsps.LoadSQSConfigFromEnv()
	app.S3 = aws.LoadS3FromEnv()
	app.DynamoDB = aws.LoadDynamoDBFromEnv()
	app.ElastiCache = aws.LoadElastiCacheFromEnv()
	app.MongoDB = mongodb.LoadConfigFromEnv()
	app.Kafka = kafka.LoadConfigFromEnv()
	app.MySQL = mysql.LoadConfigFromEnv()
	app.PostgreSQL = postgresql.LoadConfigFromEnv()
	app.Oracle = oracle.LoadConfigFromEnv()
	app.Cookie = cookie.LoadConfigFromEnv()
	app.Server = server.LoadConfigFromEnv()
	app.Metrics = metrics.LoadConfigFromEnv()
	return &app
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
		return config.LoadJSONFromConsulKV(fileName, &c).(*Config)
	}
	config.LoadJSONFile(fileName, &c)
	return &c
}
