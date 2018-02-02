package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	netContext "golang.org/x/net/context"
	"google.golang.org/appengine"

	metricscfg "github.com/NYTimes/gizmo/config/metrics"
	"github.com/NYTimes/gizmo/web"
)

// SimpleServer is a basic http Server implementation for
// serving SimpleService, JSONService or MixedService implementations.
type SimpleServer struct {
	cfg *Config

	// exit chan for graceful shutdown
	exit chan chan error

	// mux for routing
	mux Router

	svc Service

	// tracks active requests
	monitor *ActivityMonitor

	// for collecting metrics
	mets         provider.Provider
	panicCounter metrics.Counter
}

// NewSimpleServer will init the mux, exit channel and
// build the address from the given port. It will register the HealthCheckHandler
// at the given path and set up the shutDownHandler to be called on Stop().
func NewSimpleServer(cfg *Config) *SimpleServer {
	if cfg == nil {
		cfg = &Config{}
	}
	mx := NewRouter(cfg)
	if cfg.NotFoundHandler != nil {
		mx.SetNotFoundHandler(cfg.NotFoundHandler)
	}

	mets := newMetricsProvider(cfg)
	return &SimpleServer{
		mux:          mx,
		cfg:          cfg,
		exit:         make(chan chan error),
		monitor:      NewActivityMonitor(),
		mets:         mets,
		panicCounter: mets.NewCounter("panic"),
	}
}

// ServeHTTP is SimpleServer's hook for metrics and safely executing each request.
func (s *SimpleServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AddIPToContext(r)

	// only count non-LB requests
	if r.URL.Path != s.cfg.HealthCheckPath {
		s.monitor.CountRequest()
		defer s.monitor.UncountRequest()
	}

	s.safelyExecuteRequest(w, r)
}

// UnexpectedServerError is returned with a 500 status code when SimpleServer recovers
// from a panic in a request.
var UnexpectedServerError = []byte("unexpected server error")

// executeRequestSafely will prevent a panic in a request from bringing the server down.
func (s *SimpleServer) safelyExecuteRequest(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			// register a panic'd request with our metrics
			s.panicCounter.Add(1)

			// log the panic for all the details later
			LogWithFields(r).Errorf("simple server recovered from a panic\n%v: %v", x, string(debug.Stack()))

			// give the users our deepest regrets
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write(UnexpectedServerError); err != nil {
				LogWithFields(r).Warn("unable to write response: ", err)
			}
		}
	}()

	// lookup metric name if we can
	registeredPath := r.URL.Path
	if muxr, ok := s.mux.(*GorillaRouter); ok {
		var match mux.RouteMatch
		if muxr.mux.Match(r, &match) && match.MatchErr == nil {
			tmpl, err := match.Route.GetPathTemplate()
			if err == nil {
				registeredPath = tmpl
			}
		}
	}
	TimedAndCounted(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer func() {
				if err := r.Body.Close(); err != nil {
					Log.Warn("unable to close request body: ", err)
				}
			}()
		}
		s.svc.Middleware(s.mux).ServeHTTP(w, r)
	}), registeredPath, r.Method, s.mets).ServeHTTP(w, r)
}

// Start will start the SimpleServer at it's configured address.
// If they are configured, this will start health checks and access logging.
func (s *SimpleServer) Start() error {
	healthHandler := RegisterHealthHandler(s.cfg, s.monitor, s.mux)
	s.cfg.HealthCheckPath = healthHandler.Path()

	// if expvar, register on our router

	switch s.cfg.Metrics.Type {

	case metricscfg.Expvar:
		if s.cfg.Metrics.Path == "" {
			s.cfg.Metrics.Path = "/debug/vars"
		}
		s.mux.HandleFunc("GET", s.cfg.Metrics.Path, expvarHandler)

	case metricscfg.Prometheus:
		if s.cfg.Metrics.Path == "" {
			s.cfg.Metrics.Path = "/metrics"
		}
		s.mux.HandleFunc("GET", s.cfg.Metrics.Path,
			prometheus.InstrumentHandler("prometheus", prometheus.UninstrumentedHandler()))
	}

	// if this is an App Engine setup, just run it here
	if s.cfg.appEngine {
		http.Handle("/", s)
		appengine.Main()
		return nil
	}

	wrappedHandler, err := NewAccessLogMiddleware(s.cfg.HTTPAccessLog, s)
	if err != nil {
		Log.Fatalf("unable to create http access log: %s", err)
	}

	srv := httpServer(wrappedHandler)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.HTTPPort))
	if err != nil {
		return err
	}

	l = net.Listener(TCPKeepAliveListener{l.(*net.TCPListener)})

	// add TLS if in the configs
	if s.cfg.TLSCertFile != nil && s.cfg.TLSKeyFile != nil {
		cert, err := tls.LoadX509KeyPair(*s.cfg.TLSCertFile, *s.cfg.TLSKeyFile)
		if err != nil {
			return err
		}
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"http/1.1"},
		}

		l = tls.NewListener(l, srv.TLSConfig)
	}

	go func() {
		if err := srv.Serve(l); err != nil {
			Log.Error("encountered an error while serving listener: ", err)
		}
	}()
	Log.Infof("Listening on %s", l.Addr().String())

	// join the LB
	go func() {
		exit := <-s.exit

		// let the health check clean up if it needs to
		if err := healthHandler.Stop(); err != nil {
			Log.Warn("health check Stop returned with error: ", err)
		}

		// flush any remaining metrics and close connections
		s.mets.Stop()

		// stop the listener
		exit <- l.Close()
	}()

	return nil
}

// Stop initiates the shutdown process and returns when
// the server completes.
func (s *SimpleServer) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}

// Register will accept and register SimpleServer, JSONService or MixedService implementations.
func (s *SimpleServer) Register(svcI Service) error {
	s.svc = svcI
	prefix := svcI.Prefix()
	// quick fix for backwards compatibility
	prefix = strings.TrimRight(prefix, "/")

	var (
		js  JSONService
		ss  SimpleService
		cs  ContextService
		mcs MixedContextService
	)

	switch svc := svcI.(type) {
	case MixedService:
		js = svc
		ss = svc
	case SimpleService:
		ss = svc
	case JSONService:
		js = svc
	case MixedContextService:
		mcs = svc
		cs = svc
	case ContextService:
		cs = svc
	default:
		return errors.New("services for SimpleServers must implement the SimpleService, JSONService or MixedService interfaces")
	}

	if ss != nil {
		// register all simple endpoints with our wrapper
		for path, epMethods := range ss.Endpoints() {
			for method, ep := range epMethods {
				s.mux.Handle(method, prefix+path, ep)
			}
		}
	}

	if js != nil {
		// register all JSON endpoints with our wrapper
		for path, epMethods := range js.JSONEndpoints() {
			for method, ep := range epMethods {
				s.mux.Handle(method, prefix+path, JSONToHTTP(js.JSONMiddleware(ep)))
			}
		}
	}

	if cs != nil {
		// register all context endpoints with our wrapper
		for path, epMethods := range cs.ContextEndpoints() {
			for method, ep := range epMethods {
				s.mux.Handle(method, prefix+path, ContextToHTTP(cs.ContextMiddleware(ep)))
			}
		}
	}

	if mcs != nil {
		// register all context endpoints with our wrapper
		for path, epMethods := range mcs.JSONEndpoints() {
			for method, ep := range epMethods {
				// set the function handle and register it to metrics
				s.mux.Handle(method, prefix+path, ContextToHTTP(mcs.ContextMiddleware(
					JSONContextToHTTP(mcs.JSONContextMiddleware(ep)),
				)))
			}
		}
	}

	RegisterProfiler(s.cfg, s.mux)
	return nil
}

// GetForwardedIP returns the "X-Forwarded-For" header value.
func GetForwardedIP(r *http.Request) string {
	return r.Header.Get("X-Forwarded-For")
}

// GetIP returns the IP address for the given request.
func GetIP(r *http.Request) (string, error) {
	ip, ok := web.Vars(r)["ip"]
	if ok {
		return ip, nil
	}

	// check real ip header first
	ip = r.Header.Get("X-Real-IP")
	if len(ip) > 0 {
		return ip, nil
	}

	// no nginx reverse proxy?
	// get IP old fashioned way
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("%q is not IP:port", r.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", fmt.Errorf("%q is not IP:port", r.RemoteAddr)
	}
	return userIP.String(), nil
}

// ContextKey used to create context keys.
type ContextKey int

const (
	// UserIPKey is key to set/retrieve value from context.
	UserIPKey ContextKey = 0

	// UserForwardForIPKey is key to set/retrieve value from context.
	UserForwardForIPKey ContextKey = 1
)

// ContextWithUserIP returns new context with user ip address.
func ContextWithUserIP(ctx netContext.Context, r *http.Request) netContext.Context {
	ip, err := GetIP(r)
	if err != nil {
		LogWithFields(r).Warningf("unable to get IP: %s", err)
		return ctx
	}
	return netContext.WithValue(ctx, UserIPKey, ip)
}

// ContextWithForwardForIP returns new context with forward for ip.
func ContextWithForwardForIP(ctx netContext.Context, r *http.Request) netContext.Context {
	ip := GetForwardedIP(r)
	if len(ip) > 0 {
		ctx = netContext.WithValue(ctx, UserForwardForIPKey, ip)
	}

	return ctx
}
