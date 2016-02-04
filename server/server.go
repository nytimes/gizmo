package server

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
	"github.com/rcrowley/go-metrics"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/web"
	"github.com/NYTimes/logrotate"
)

// Version is meant to be set with the current package version at build time.
var Version string

// Server is the basic interface that defines what to expect from any server.
type Server interface {
	Register(Service) error
	Start() error
	Stop() error
}

var (
	// Name is used for status and logging.
	Name = "nyt-awesome-go-server"
	// Log is the global logger for the server. It will take care of logrotate
	// and it can accept 'fields' to include with each log line: see LogWithFields(r).
	Log = logrus.New()
	// server is what's used in the global server funcs in the package.
	server Server
	// MaxHeaderBytes is used by the http server to limit the size of request headers.
	// This may need to be increased if accepting cookies from the public.
	maxHeaderBytes = 1 << 20
	// JSONContentType is the content type that will be used for JSONEndpoints.
	// It will default to the web.JSONContentType value.
	jsonContentType = web.JSONContentType
)

// Init will set up our name, logging, healthchecks and parse flags. If DefaultServer isn't set,
// this func will set it to a `SimpleServer` listening on `Config.Server.HTTPPort`.
func Init(name string, scfg *config.Server) {
	// generate a unique ID for the server
	id, _ := uuid.NewV4()
	Name = name + "-" + Version + "-" + id.String()

	// if no config given, attempt to pull one from
	// the environment.
	if scfg == nil {
		// allow the default config to be overridden by CLI
		flag.Parse()
		cfg := config.NewConfig(*config.ConfigLocationCLI)
		config.SetServerOverrides(cfg.Server)
		scfg = cfg.Server
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

	// setup app logging
	if scfg.Log != "" {
		lf, err := logrotate.NewFile(scfg.Log)
		if err != nil {
			Log.Fatalf("unable to access log file: %s", err)
		}
		Log.Out = lf
		// json output when writing to file
		Log.Formatter = &logrus.JSONFormatter{}
	} else {
		Log.Out = os.Stderr
	}
	SetLogLevel(scfg)

	server = NewServer(scfg)
}

// Register will add a new Service to the DefaultServer.
func Register(svc Service) error {
	return server.Register(svc)
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

// ContextFields will take a request and convert a context map to logrus Fields.
func ContextFields(r *http.Request) map[string]interface{} {
	fields := map[string]interface{}{}
	for k, v := range context.GetAll(r) {
		strK := fmt.Sprintf("%+v", k)
		// gorilla.mux adds the route to context.
		// we want to remove it for now
		if strK == "1" {
			continue
		}
		// gorilla puts mux vars here, we want to give a better label
		if strK == "0" {
			strK = "muxvars"
		}
		fields[strK] = fmt.Sprintf("%#+v", v)
	}
	fields["path"] = r.URL.Path
	fields["rawquery"] = r.URL.RawQuery

	return fields
}

// NewServer will inspect the config and generate
// the appropriate Server implementation.
func NewServer(cfg *config.Server) Server {
	switch cfg.ServerType {
	case "simple":
		return NewSimpleServer(cfg)
	case "rpc":
		return NewRPCServer(cfg)
	default:
		return NewSimpleServer(cfg)
	}
}

// NewHealthCheckHandler will inspect the config to generate
// the appropriate HealthCheckHandler.
func NewHealthCheckHandler(cfg *config.Server) HealthCheckHandler {
	switch cfg.HealthCheckType {
	case "simple":
		return NewSimpleHealthCheck(cfg.HealthCheckPath)
	case "esx":
		return NewESXHealthCheck()
	default:
		return NewSimpleHealthCheck("/status.txt")
	}
}

// RegisterProfiler will add handlers for pprof endpoints if
// the config has them enabled.
func RegisterProfiler(cfg *config.Server, mx *mux.Router) {
	if !cfg.EnablePProf {
		return
	}
	mx.HandleFunc("/debug/pprof/", pprof.Index)
	mx.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mx.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mx.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	mx.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mx.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mx.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mx.Handle("/debug/pprof/block", pprof.Handler("block"))
}

// RegisterHealthHandler will create a new HealthCheckHandler from the
// given config and add a handler to the given router.
func RegisterHealthHandler(cfg *config.Server, monitor *ActivityMonitor, mx *mux.Router) HealthCheckHandler {
	// register health check
	hch := NewHealthCheckHandler(cfg)
	err := hch.Start(monitor)
	if err != nil {
		Log.Fatal("unable to start the HealthCheckHandler: ", err)
	}
	mx.Handle(hch.Path(), hch)
	return hch
}

// StartServerMetrics will start emitting metrics to the provided
// registry (nil means the DefaultRegistry) if a Graphite host name
// is given in the config.
func StartServerMetrics(cfg *config.Server, registry metrics.Registry) {
	if registry == nil {
		registry = metrics.DefaultRegistry
	}
	if cfg.GraphiteHost == "" {
		return
	}
	Log.Infof("connecting to graphite host: %s", cfg.GraphiteHost)
	addr, err := net.ResolveTCPAddr("tcp", cfg.GraphiteHost)
	if err != nil {
		Log.Warnf("unable to resolve graphite host: %s", err)
	}
	go graphite.Graphite(registry, 30*time.Second, MetricsRegistryName(), addr)
}

// RegisterAccessLogger will wrap a logrotate-aware Apache-style access log handler
// around the given handler if an access log location is provided by the config.
func RegisterAccessLogger(cfg *config.Server, handler http.Handler) http.Handler {
	if len(cfg.HTTPAccessLog) == 0 {
		return handler
	}

	lf, err := logrotate.NewFile(cfg.HTTPAccessLog)
	if err != nil {
		Log.Fatalf("unable to access http access log file: %s", err)
	}
	return handlers.CombinedLoggingHandler(lf, handler)
}

// MetricsRegistryName returns "apps.{hostname prefix}", which is
// the convention used in NYT ESX environment.
func MetricsRegistryName() string {
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
func SetLogLevel(scfg *config.Server) {
	switch scfg.LogLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
