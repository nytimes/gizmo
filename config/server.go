package config

import (
	"github.com/NYTimes/logrotate"
	"github.com/gorilla/handlers"
	"io"
	"net/http"
	"os"

	"github.com/rcrowley/go-metrics"
)

// Server holds info required to configure a gizmo server.Server.
type Server struct {
	// Server will tell the server package which type of server to init. If
	// empty, this will default to 'simple'.
	ServerType string `envconfig:"GIZMO_SERVER_TYPE"`
	// HealthCheckType is used by server to init the proper HealthCheckHandler.
	// If empty, this will default to 'simple'.
	HealthCheckType string `envconfig:"GIZMO_HEALTH_CHECK_TYPE"`
	// RouterType is used by the server to init the proper Router implementation.
	// If empty, this will default to 'gorilla'.
	RouterType string `envconfig:"GIZMO_ROUTER_TYPE"`
	// HealthCheckPath is used by server to init the proper HealthCheckHandler.
	// If empty, this will default to '/status.txt'.
	HealthCheckPath string `envconfig:"GIZMO_HEALTH_CHECK_PATH"`
	// JSONContentType can be used to override the default JSONContentType.
	JSONContentType *string `envconfig:"GIZMO_JSON_CONTENT_TYPE"`
	// MaxHeaderBytes can be used to override the default MaxHeaderBytes (1<<20).
	MaxHeaderBytes *int `envconfig:"GIZMO_JSON_CONTENT_TYPE"`
	// ReadTimeout can be used to override the default http server timeout of 10s.
	// The string should be formatted like a time.Duration string.
	ReadTimeout *string `envconfig:"GIZMO_READ_TIMEOUT"`
	// WriteTimeout can be used to override the default http server timeout of 10s.
	// The string should be formatted like a time.Duration string.
	WriteTimeout *string `envconfig:"GIZMO_WRITE_TIMEOUT"`
	// GOMAXPROCS can be used to override the default GOMAXPROCS (runtime.NumCPU).
	GOMAXPROCS *int `envconfig:"GIZMO_SERVER_GOMAXPROCS"`
	// HTTPAccessLog is the location of the http access log. If it is empty,
	// no access logging will be done.
	HTTPAccessLog *string `envconfig:"HTTP_ACCESS_LOG"`
	// RPCAccessLog is the location of the RPC access log. If it is empty,
	// no access logging will be done.
	RPCAccessLog *string `envconfig:"RPC_ACCESS_LOG"`
	// HTTPPort is the port the server implementation will serve HTTP over.
	HTTPPort int `envconfig:"HTTP_PORT"`
	// RPCPort is the port the server implementation will serve RPC over.
	RPCPort int `envconfig:"RPC_PORT"`
	// Log is the path to the application log.
	Log string `envconfig:"APP_LOG"`
	// LogLevel will override the default log level of 'info'.
	LogLevel string `envconfig:"APP_LOG_LEVEL"`
	// Enable pprof Profiling. Off by default.
	EnablePProf bool `envconfig:"ENABLE_PPROF"`
	// GraphiteHost should be the host and port of an available graphite cluster.
	// If not set, the server will not emit metrics.
	GraphiteHost string `envconfig:"GRAPHITE_HOST"`
	// TLSCertFile is an optional string for enabling TLS in simple servers.
	TLSCertFile *string `envconfig:"TLS_CERT"`
	// TLSKeyFile is an optional string for enabling TLS in simple servers.
	TLSKeyFile *string `envconfig:"TLS_KEY"`
	// NotFoundHandler will override the default server NotfoundHandler if set.
	NotFoundHandler http.Handler
	// MetricsRegistry will override the default server metrics registry if set.
	MetricsRegistry metrics.Registry
}

// LoadServerFromEnv will attempt to load a Server object
// from environment variables. If not populated, nil
// is returned.
func LoadServerFromEnv() *Server {
	var server Server
	LoadEnvConfig(&server)
	if server.HTTPPort != 0 || server.RPCPort != 0 ||
		server.HTTPAccessLog != nil || server.RPCAccessLog != nil ||
		server.HealthCheckType != "" || server.HealthCheckPath != "" {
		return &server
	}
	return nil
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
	if *logLocation == "stdout" {
		lw = os.Stdout
	} else {
		lw, err = logrotate.NewFile(*logLocation)
		if err != nil {
			return nil, err
		}
	}
	return handlers.CombinedLoggingHandler(lw, handler), nil
}
