package kithttp

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	gserver "github.com/NYTimes/gizmo/server"
	httptransport "github.com/go-kit/kit/transport/http"
)

// TODO(jprobinson): USE kit/log
// TODO(jprobinson): USE kit/metrics/prometheus
// TODO(jprobinson): built in stackdriver error reporting
// TODO(jprobinson): built in stackdriver tracing (sampling)

var (
	// maxHeaderBytes is used by the http server to limit the size of request headers.
	// This may need to be increased if accepting cookies from the public.
	maxHeaderBytes = 1 << 20
	// readTimeout is used by the http server to set a maximum duration before
	// timing out read of the request. The default timeout is 10 seconds.
	readTimeout = 10 * time.Second
	// writeTimeout is used by the http server to set a maximum duration before
	// timing out write of the response. The default timeout is 10 seconds.
	writeTimeout = 10 * time.Second
	// idleTimeout is used by the http server to set a maximum duration for
	// keep-alive connections.
	idleTimeout = 120 * time.Second
)

type contextKey int

const (
	ContextKeyInboundAppID contextKey = iota
	// key to set/retrieve URL params from a request context.
	varsKey
)

var defaultOpts = []httptransport.ServerOption{
	httptransport.ServerBefore(
		// populate context with helpful keys
		httptransport.PopulateRequestContext,
	),
}

// Run will use environment variables to configure the
// server, register the given Service and start up the
// HTTP server.
// This will block until the server shuts down.
func Run(service Service) error {
	// allow the default config to be overridden by CLI
	scfg := gserver.LoadConfigFromEnv()
	gserver.SetConfigOverrides(scfg)

	if scfg.GOMAXPROCS != nil {
		runtime.GOMAXPROCS(*scfg.GOMAXPROCS)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if scfg.MaxHeaderBytes != nil {
		maxHeaderBytes = *scfg.MaxHeaderBytes
	}

	if scfg.ReadTimeout != nil {
		tReadTimeout, err := time.ParseDuration(*scfg.ReadTimeout)
		if err != nil {
			panic("invalid server ReadTimeout: " + err.Error())
		}
		readTimeout = tReadTimeout
	}

	if scfg.IdleTimeout != nil {
		tIdleTimeout, err := time.ParseDuration(*scfg.IdleTimeout)
		if err != nil {
			panic("invalid server IdleTimeout: " + err.Error())
		}
		idleTimeout = tIdleTimeout
	}

	if scfg.WriteTimeout != nil {
		tWriteTimeout, err := time.ParseDuration(*scfg.WriteTimeout)
		if err != nil {
			panic("invalid server WriteTimeout: " + err.Error())
		}
		writeTimeout = tWriteTimeout
	}

	srvr := newServer(*scfg, service.RouterOptions()...)
	err := srvr.Register(service)
	if err != nil {
		panic("unable to register service: " + err.Error())
	}

	srvr.Logger.Log("Starting new server")
	if err := srvr.Start(); err != nil {
		return err
	}

	// parse address for host, port
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	srvr.Logger.Log("Received signal %s", <-ch)
	return srvr.Stop()
}
