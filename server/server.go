package server

import (
	"errors"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/NYTimes/logrotate"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/sirupsen/logrus"
)

// Version is meant to be set with the current package version at build time.
var Version string

// Server is the basic interface that defines what to expect from any server.
type Server interface {
	Register(*mux.Router) error
	Start() error
	Stop() error
}

var (
	// ErrMultiRegister occurs when a Register method is called multiple times
	ErrMultiRegister = errors.New("register method has been called multiple times")

	// Name is used for status and logging.
	Name = "nyt-awesome-go-server"
	// Log is the global logger for the server. It will take care of logrotate
	// and it can accept 'fields' to include with each log line: see LogWithFields(r).
	Log = logrus.New()
	// server is what's used in the global server funcs in the package.
	server Server
	// maxHeaderBytes is used by the http server to limit the size of request headers.
	// This may need to be increased if accepting cookies from the public.
	maxHeaderBytes = 1 << 20
	// readTimeout is used by the http server to set a maximum duration before
	// timing out read of the request. The default timeout is 10 seconds.
	readTimeout = 10 * time.Second
	// writeTimeout is used by the http server to set a maximum duration before
	// timing out write of the response. The default timeout is 10 seconds.
	writeTimeout = 10 * time.Second
	// jsonContentType is the content type that will be used for JSONEndpoints.
	// It will default to the JSONContentType value.
	jsonContentType = JSONContentType
	// idleTimeout is used by the http server to set a maximum duration for
	// keep-alive connections.
	idleTimeout = 120 * time.Second
)

// Init will set up our name, logging, healthchecks and parse flags. If DefaultServer isn't set,
// this func will set it to a `SimpleServer` listening on `Config.HTTPPort`.
func Init(name string, scfg *Config) {
	// generate a unique ID for the server
	id, _ := uuid.NewV4()
	Name = name + "-" + Version + "-" + id.String()

	// if no config given, attempt to pull one from
	// the environment.
	if scfg == nil {
		// allow the default config to be overridden by CLI
		scfg = LoadConfigFromEnv()
		SetConfigOverrides(scfg)
	}

	if scfg.GOMAXPROCS != nil {
		runtime.GOMAXPROCS(*scfg.GOMAXPROCS)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if scfg.JSONContentType != nil {
		jsonContentType = *scfg.JSONContentType
	}

	if scfg.MaxHeaderBytes != nil {
		maxHeaderBytes = *scfg.MaxHeaderBytes
	}

	if scfg.ReadTimeout != nil {
		tReadTimeout, err := time.ParseDuration(*scfg.ReadTimeout)
		if err != nil {
			Log.Fatal("invalid server ReadTimeout: ", err)
		}
		readTimeout = tReadTimeout
	}

	if scfg.IdleTimeout != nil {
		tIdleTimeout, err := time.ParseDuration(*scfg.IdleTimeout)
		if err != nil {
			Log.Fatal("invalid server IdleTimeout: ", err)
		}
		idleTimeout = tIdleTimeout
	}

	if scfg.WriteTimeout != nil {
		tWriteTimeout, err := time.ParseDuration(*scfg.WriteTimeout)
		if err != nil {
			Log.Fatal("invalid server WriteTimeout: ", err)
		}
		writeTimeout = tWriteTimeout
	}

	// setup app logging
	if scfg.Log != "" {
		lf, err := logrotate.NewFile(scfg.Log)
		if err != nil {
			Log.Fatalf("unable to access log file: %s", err)
		}
		Log.Out = lf

		// json output when writing to file by default
		if scfg.LogJSONFormat == nil {
			Log.Formatter = &logrus.JSONFormatter{}
		}

	} else {
		Log.Out = os.Stderr
	}

	// override default JSON settings
	if scfg.LogJSONFormat != nil && *scfg.LogJSONFormat {
		Log.Formatter = &logrus.JSONFormatter{}
	}

	SetLogLevel(scfg)

	server = NewServer(scfg)
}

// Register will add a new Service to the DefaultServer.
func Register(r *mux.Router) error {
	return server.Register(r)
}

// Run will start the DefaultServer and set it up to Stop()
// on a kill signal.
func Run() error {
	Log.Infof("Starting new %s server", Name)
	if err := server.Start(); err != nil {
		return err
	}

	// parse address for host, port
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	Log.Infof("Received signal %s", <-ch)
	return Stop()
}

// Stop will stop the default server.
func Stop() error {
	Log.Infof("Stopping %s server", Name)
	return server.Stop()
}

// LogWithFields will feed any request context into a logrus Entry.
func LogWithFields(r *http.Request) *logrus.Entry {
	return Log.WithFields(ContextFields(r))
}

// NewServer will inspect the config and generate
// the appropriate Server implementation.
func NewServer(cfg *Config) Server {
	return NewSimpleServer(cfg)
}

// NewHealthCheckHandler will inspect the config to generate
// the appropriate HealthCheckHandler.
func NewHealthCheckHandler(cfg *Config) (HealthCheckHandler, error) {
	// default the status path if not set
	if cfg.HealthCheckPath == "" {
		cfg.HealthCheckPath = "/status.txt"
	}
	switch cfg.HealthCheckType {
	case "simple":
		return NewSimpleHealthCheck(cfg.HealthCheckPath), nil
	case "custom":
		if cfg.CustomHealthCheckHandler == nil {
			return nil, errors.New("health check type is set to 'custom', but no Config.CustomHealthCheckHandler provided")
		}
		return NewCustomHealthCheck(cfg.HealthCheckPath, cfg.CustomHealthCheckHandler), nil
	default:
		return NewSimpleHealthCheck(cfg.HealthCheckPath), nil
	}
}

// RegisterProfiler will add handlers for pprof endpoints if
// the config has them enabled.
func RegisterProfiler(cfg *Config, mx *mux.Router) {
	if !cfg.EnablePProf {
		return
	}
	mx.HandleFunc("/debug/pprof/", pprof.Index).Methods("GET")
	mx.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Methods("GET")
	mx.HandleFunc("/debug/pprof/profile", pprof.Profile).Methods("GET")
	mx.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Methods("GET")
	mx.HandleFunc("/debug/pprof/trace", pprof.Trace).Methods("GET")

	// Manually add support for paths linked to by index page at /debug/pprof/
	mx.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine")).Methods("GET")
	mx.Handle("/debug/pprof/heap", pprof.Handler("heap")).Methods("GET")
	mx.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate")).Methods("GET")
	mx.Handle("/debug/pprof/block", pprof.Handler("block")).Methods("GET")
}

// RegisterHealthHandler will create a new HealthCheckHandler from the
// given config and add a handler to the given router.
func RegisterHealthHandler(cfg *Config, monitor *ActivityMonitor, mx *mux.Router) HealthCheckHandler {
	// register health check
	hch, err := NewHealthCheckHandler(cfg)
	if err != nil {
		Log.Fatal("unable to configure the HealthCheckHandler: ", err)
	}
	err = hch.Start(monitor)
	if err != nil {
		Log.Fatal("unable to start the HealthCheckHandler: ", err)
	}
	mx.Handle(hch.Path(), hch).Methods("GET", "HEAD")
	return hch
}

// MetricsNamespace returns "apps.{hostname prefix}", which is
// the convention used in NYT ESX environment.
func MetricsNamespace() string {
	// get only server base name
	name, _ := os.Hostname()
	name = strings.SplitN(name, ".", 2)[0]
	// set it up to be paperboy.servername
	name = strings.Replace(name, "-", ".", 1)
	// add the 'apps' prefix to keep things neat
	return "apps." + name
}

// SetLogLevel will set the appropriate logrus log level
// given the server config.
func SetLogLevel(scfg *Config) {
	if lvl, err := logrus.ParseLevel(scfg.LogLevel); err != nil {
		Log.Level = logrus.InfoLevel
	} else {
		Log.Level = lvl
	}
}
