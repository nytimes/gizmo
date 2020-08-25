package server

import (
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/logrotate"
	"github.com/gorilla/handlers"
	"github.com/kelseyhightower/envconfig"
)

// Config holds info required to configure a gizmo server.Server.
type Config struct {
	// HealthCheckType is used by server to init the proper HealthCheckHandler.
	// If empty, this will default to 'simple'.
	HealthCheckType string `envconfig:"GIZMO_HEALTH_CHECK_TYPE"`
	// HealthCheckPath is used by server to init the proper HealthCheckHandler.
	// If empty, this will default to '/status.txt'.
	HealthCheckPath string `envconfig:"GIZMO_HEALTH_CHECK_PATH"`
	// CustomHealthCheckHandler will be used if HealthCheckType is set with "custom".
	CustomHealthCheckHandler http.Handler

	// RouterType is used by the server to init the proper Router implementation.
	// The current available types are 'gorilla' to use the Gorilla tool kit mux and
	// 'stdlib' to use the http package's ServeMux.
	// If empty, this will default to 'gorilla'.
	// NOTE: If 'stdlib' is used, Prometheus monitoring will be disabled for performance
	// reasons.
	RouterType string `envconfig:"GIZMO_ROUTER_TYPE"`

	// JSONContentType can be used to override the default JSONContentType.
	JSONContentType *string `envconfig:"GIZMO_JSON_CONTENT_TYPE"`
	// MaxHeaderBytes can be used to override the default MaxHeaderBytes (1<<20).
	MaxHeaderBytes *int `envconfig:"GIZMO_MAX_HEADER_BYTES"`
	// ReadTimeout can be used to override the default http server timeout of 10s.
	// The string should be formatted like a time.Duration string.
	ReadTimeout *string `envconfig:"GIZMO_READ_TIMEOUT"`
	// WriteTimeout can be used to override the default http server timeout of 10s.
	// The string should be formatted like a time.Duration string.
	WriteTimeout *string `envconfig:"GIZMO_WRITE_TIMEOUT"`
	// IdleTimeout can be used to override the default http server timeout of 120s.
	// The string should be formatted like a time.Duration string. This
	// feature is supported only on Go 1.8+.
	IdleTimeout *string `envconfig:"GIZMO_IDLE_TIMEOUT"`

	// GOMAXPROCS can be used to override the default GOMAXPROCS (runtime.NumCPU).
	GOMAXPROCS *int `envconfig:"GIZMO_SERVER_GOMAXPROCS"`

	// HTTPAccessLog is the location of the http access log. If it is empty,
	// no access logging will be done.
	HTTPAccessLog *string `envconfig:"HTTP_ACCESS_LOG"`
	// RPCAccessLog is the location of the RPC access log. If it is empty,
	// no access logging will be done.
	RPCAccessLog *string `envconfig:"RPC_ACCESS_LOG"`

	// HTTPAddr is the address the server implementation will bind to.
	HTTPAddr string `envconfig:"HTTP_ADDR"`
	// HTTPPort is the port the server implementation will serve HTTP over.
	HTTPPort int `envconfig:"HTTP_PORT"`
	// RPCPort is the port the server implementation will serve RPC over.
	RPCPort int `envconfig:"RPC_PORT"`

	// Log is the path to the application log.
	Log string `envconfig:"APP_LOG"`
	// LogLevel will override the default log level of 'info'.
	LogLevel string `envconfig:"APP_LOG_LEVEL"`
	// LogJSONFormat will override the default JSON formatting logic on the server.
	// By default the Server will log in a JSON format only if the Log field
	// is defined.
	// If this field is set and true, the logrus JSONFormatter will be used.
	LogJSONFormat *bool `envconfig:"APP_LOG_JSON_FMT"`

	// TLSCertFile is an optional string for enabling TLS in simple servers.
	TLSCertFile *string `envconfig:"TLS_CERT"`
	// TLSKeyFile is an optional string for enabling TLS in simple servers.
	TLSKeyFile *string `envconfig:"TLS_KEY"`

	// NotFoundHandler will override the default server NotfoundHandler if set.
	NotFoundHandler http.Handler

	// Enable pprof Profiling. Off by default.
	EnablePProf bool `envconfig:"ENABLE_PPROF"`

	// MetricsNamespace is used by prometheus.
	MetricsNamespace string `envconfig:"METRICS_NAMESPACE"`
	// MetricsSubsystem is used by prometheus.
	MetricsSubsystem string `envconfig:"METRICS_SUBSYSTEM"`
	// MetricsPath is where the prometheus endpoint will be registered.
	MetricsPath string `envconfig:"METRICS_PATH"`
}

// LoadConfigFromEnv will attempt to load a Server object
// from environment variables. If not populated, nil
// is returned.
func LoadConfigFromEnv() *Config {
	var server Config
	envconfig.Process("", &server)
	return &server
}

// NewAccessLogMiddleware will wrap a logrotate-aware Apache-style access log handler
// around the given http.Handler if an access log location is provided by the config,
// or optionally send access logs to stdout.
func NewAccessLogMiddleware(logLocation *string, handler http.Handler) (http.Handler, error) {
	if logLocation == nil {
		return handler, nil
	}
	var lw io.Writer
	var err error
	switch *logLocation {
	case "stdout":
		lw = os.Stdout
	default:
		lw, err = logrotate.NewFile(*logLocation)
		if err != nil {
			return nil, err
		}
	}
	return handlers.CombinedLoggingHandler(lw, handler), nil
}

// SetConfigOverrides will check the *CLI variables for any values
// and override the values in the given config if they are set.
// If LogCLI is set to "dev", the given `Log` pointer will be set to an
// empty string.
func SetConfigOverrides(c *Config) {
	// HTTPAccessLogCLI is a pointer to the value of the '-http-access-log' command line flag. It is meant to
	// declare an access log location for HTTP services.
	HTTPAccessLogCLI := flag.String("http-access-log", "", "HTTP access log location")
	// RPCAccessLogCLI is a pointer to the value of the '-rpc-access-log' command line flag. It is meant to
	// declare an acces log location for RPC services.
	RPCAccessLogCLI := flag.String("rpc-access-log", "", "RPC access log location")
	// HTTPPortCLI is a pointer to the value for the '-http' flag. It is meant to declare the port
	// number to serve HTTP services.
	HTTPPortCLI := flag.Int("http", 0, "Port to run an HTTP server on")
	// RPCPortCLI is a pointer to the value for the '-rpc' flag. It is meant to declare the port
	// number to serve RPC services.
	RPCPortCLI := flag.Int("rpc", 0, "Port to run an RPC server on")

	config.SetLogOverride(&c.Log)

	if *HTTPAccessLogCLI != "" {
		c.HTTPAccessLog = HTTPAccessLogCLI
	}

	if *RPCAccessLogCLI != "" {
		c.RPCAccessLog = RPCAccessLogCLI
	}

	if *HTTPPortCLI > 0 {
		c.HTTPPort = *HTTPPortCLI
	}

	if *RPCPortCLI > 0 {
		c.RPCPort = *RPCPortCLI
	}
}
