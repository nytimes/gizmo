package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nytimes/gizmo/config"
	"github.com/rcrowley/go-metrics"
)

// SimpleServer is a basic http Server implementation for
// serving SimpleService, JSONService or MixedService implementations.
type SimpleServer struct {
	cfg *config.Server

	// exit chan for graceful shutdown
	exit chan chan error

	// mux for routing
	mux *mux.Router

	// tracks active requests
	monitor *ActivityMonitor
}

// NewSimpleServer will init the mux, exit channel and
// build the address from the given port. It will register the HealthCheckHandler
// at the given path and set up the shutDownHandler to be called on Stop().
func NewSimpleServer(cfg *config.Server) *SimpleServer {
	if cfg == nil {
		cfg = &config.Server{}
	}
	mx := mux.NewRouter()
	if cfg.NotFoundHandler != nil {
		mx.NotFoundHandler = cfg.NotFoundHandler
	}
	return &SimpleServer{
		mux:     mx,
		cfg:     cfg,
		exit:    make(chan chan error),
		monitor: NewActivityMonitor(),
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
			errCntr := metrics.GetOrRegisterCounter("PANIC", metrics.DefaultRegistry)
			errCntr.Inc(1)

			// log the panic for all the details later
			LogWithFields(r).Errorf("simple server recovered from a panic\n%v: %v", x, string(debug.Stack()))

			// give the users our deepest regrets
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(UnexpectedServerError)
		}
	}()

	// hand the request off to gorilla
	s.mux.ServeHTTP(w, r)
}

// Start will start the SimpleServer at it's configured address.
// If they are configured, this will start emitting metrics to Graphite,
// register profiling, health checks and access logging.
func (s *SimpleServer) Start() error {

	StartServerMetrics(s.cfg)

	healthHandler := RegisterHealthHandler(s.cfg, s.monitor, s.mux)
	s.cfg.HealthCheckPath = healthHandler.Path()

	srv := http.Server{
		Handler:        RegisterAccessLogger(s.cfg, s),
		MaxHeaderBytes: maxHeaderBytes,
	}

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

	go srv.Serve(l)
	Log.Infof("Listening on %s", l.Addr().String())

	// join the LB
	go func() {
		exit := <-s.exit

		// let the health check clean up if it needs to
		healthHandler.Stop()

		// stop the listener
		exit <- l.Close()
	}()

	return nil
}

// Stop initiates the shutdown process and returns when
// the server completes
func (s *SimpleServer) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}

func metricName(prefix, path, method string) string {
	// combine and trim prefix
	fullpath := strings.TrimPrefix(prefix+path, "/")
	// replace slashes
	fullpath = strings.Replace(fullpath, "/", "-", -1)
	// replace periods
	fullpath = strings.Replace(fullpath, ".", "-", -1)
	return fmt.Sprintf("routes.%s-%s", fullpath, method)
}

// Register will accept and register SimpleServer, JSONService or MixedService implementations.
func (s *SimpleServer) Register(svcI Service) error {
	prefix := svcI.Prefix()
	sr := s.mux.PathPrefix(prefix).Subrouter()

	var (
		js JSONService
		ss SimpleService
	)

	switch svc := svcI.(type) {
	case MixedService:
		js = svc
		ss = svc
	case SimpleService:
		ss = svc
	case JSONService:
		js = svc
	default:
		return errors.New("services for SimpleServers must implement the SimpleService, JSONService or MixedService interfaces")
	}

	if ss != nil {
		// register all simple endpoints with our wrapper
		for path, epMethods := range ss.Endpoints() {
			for method, ep := range epMethods {
				endpointName := metricName(prefix, path, method)
				// set the function handle and register is to metrics
				sr.Handle(path, Timed(CountedByStatusXX(
					func(ep http.HandlerFunc, ss SimpleService) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							// is it worth it to always close this?
							if r.Body != nil {
								defer r.Body.Close()
							}

							// call the func and return err or not
							ss.Middleware(ep).ServeHTTP(w, r)
						})
					}(ep, ss),
					endpointName+".STATUS-COUNT", metrics.DefaultRegistry),
					endpointName+".DURATION", metrics.DefaultRegistry),
				).Methods(method)
			}
		}
	}

	if js != nil {
		// register all JSON endpoints with our wrapper
		for path, epMethods := range js.JSONEndpoints() {
			for method, ep := range epMethods {
				endpointName := metricName(prefix, path, method)
				// set the function handle and register is to metrics
				sr.Handle(path, Timed(CountedByStatusXX(
					js.Middleware(JSONToHTTP(js.JSONMiddleware(ep))),
					endpointName+".STATUS-COUNT", metrics.DefaultRegistry),
					endpointName+".DURATION", metrics.DefaultRegistry),
				).Methods(method)
			}
		}
	}

	RegisterProfiler(s.cfg, s.mux)
	return nil
}

// AddIPToContext will attempt to pull an IP address out of the request and
// set it into a gorilla context
func AddIPToContext(r *http.Request) {
	ip, err := GetIP(r)
	if err != nil {
		LogWithFields(r).Warningf("unable to get IP: %s", err)
	} else {
		context.Set(r, "ip", ip)
	}

	if ip = GetForwardedIP(r); len(ip) > 0 {
		context.Set(r, "forward-for-ip", ip)
	}
}

// GetForwardedIP returns the "X-Forwarded-For" header value
func GetForwardedIP(r *http.Request) string {
	return r.Header.Get("X-Forwarded-For")
}

// GetIP returns the IP address for the given request.
func GetIP(r *http.Request) (string, error) {
	ip, ok := mux.Vars(r)["ip"]
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
