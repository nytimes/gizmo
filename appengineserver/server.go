package appengineserver

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/healthcheck"
	"github.com/NYTimes/gizmo/web"
)

// Server is the basic interface that defines what to expect from any server.
type Server interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Register(Service) error
	Start() error
	Stop() error
}

var (
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
func Init(scfg *config.Server, service Service) {
	InitVM(scfg)
	err := server.Register(service)
	if err != nil {
		panic("unable to register service: " + err.Error())
	}
	http.Handle("/", server)
}

func InitVM(scfg *config.Server) {
	// if no config given, attempt to pull one from
	// the environment.
	if scfg == nil {
		scfg = config.LoadServerFromEnv()
		// in case no env vars found
		if scfg == nil {
			scfg = &config.Server{}
		}
	}

	if scfg.JSONContentType != nil {
		jsonContentType = *scfg.JSONContentType
	}

	if scfg.MaxHeaderBytes != nil {
		maxHeaderBytes = *scfg.MaxHeaderBytes
	}

	server = NewServer(scfg)
}

// Register will add a new Service to the DefaultServer.
func Register(svc Service) error {
	return server.Register(svc)
}

// Run will start the DefaultServer and set it up to Stop()
// on a kill signal.
func Run() error {
	if err := server.Start(); err != nil {
		return err
	}

	// parse address for host, port
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	return Stop()
}

// Stop will stop the default server.
func Stop() error {
	return server.Stop()
}

// NewHealthCheckHandler will inspect the config to generate
// the appropriate HealthCheckHandler.
func NewHealthCheckHandler(cfg *config.Server) healthcheck.Handler {
	switch cfg.HealthCheckType {
	case "simple":
		return healthcheck.NewSimple(cfg.HealthCheckPath)
	case "esx":
		return healthcheck.NewESX()
	default:
		return healthcheck.NewSimple("/_ah/health")
	}
}

// NewServer will inspect the config and generate
// the appropriate Server implementation.
func NewServer(cfg *config.Server) Server {
	switch cfg.ServerType {
	case "simple":
		return NewSimpleServer(cfg)
		//	case "rpc":
		//		return NewRPCServer(cfg)
	default:
		return NewSimpleServer(cfg)
	}
}
