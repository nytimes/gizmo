package kit

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"cloud.google.com/go/errorreporting"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	ocontext "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server encapsulates all logic for registering and running a gizmo kit server.
type Server struct {
	logger   log.Logger
	logClose func() error

	errs *errorreporting.Client

	mux Router

	cfg Config

	svc Service

	svr  *http.Server
	gsvr *grpc.Server

	// exit chan for graceful shutdown
	exit chan chan error
}

type contextKey int

const (
	// key to set/retrieve URL params from a request context.
	varsKey contextKey = iota
	// key for logger
	logKey

	// ContextKeyCloudTraceContext is a context key for storing and retrieving the
	// inbound 'x-cloud-trace-context' header. This server will automatically look for
	// and inject the value into the request context. If in the App Engine environment
	// this will be used to enable combined access and application logs.
	ContextKeyCloudTraceContext
)

// NewServer will create a new kit server for the given Service.
//
// Generally, users should only use the 'Run' function to start a server and use this
// function within tests so they may call ServeHTTP.
func NewServer(svc Service) *Server {
	// load config from environment with defaults set
	cfg := loadConfig()

	ropts := svc.HTTPRouterOptions()
	// default the router if none set
	if len(ropts) == 0 {
		ropts = append(ropts, RouterSelect(""))
	}
	var r Router
	for _, opt := range ropts {
		r = opt(r)
	}

	// check if we're running on GAE via env variables
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	serviceID := os.Getenv("GAE_SERVICE")
	svcVersion := os.Getenv("GAE_VERSION")

	var (
		err      error
		lg       log.Logger
		logClose func() error
		errs     *errorreporting.Client
	)
	// use the version variable to determine if we're in the GAE environment
	if svcVersion == "" {
		lg = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	} else {
		ctx := context.Background()
		lg, logClose, err = newAppEngineLogger(ctx,
			projectID, serviceID, svcVersion)
		if err != nil {
			stdlog.Fatalf("unable to start up app engine logger: %s", err)
		}

		errs, err = errorreporting.NewClient(ctx, projectID, errorreporting.Config{
			ServiceName:    serviceID,
			ServiceVersion: svcVersion,
			OnError: func(err error) {
				lg.Log("error", err,
					"message", "error reporting client encountered an error")
			},
		})
	}

	s := &Server{
		cfg:      cfg,
		mux:      r,
		exit:     make(chan chan error),
		logger:   lg,
		logClose: logClose,
		errs:     errs,
	}
	s.svr = &http.Server{
		Handler:        s,
		Addr:           fmt.Sprintf(":%d", cfg.HTTPPort),
		MaxHeaderBytes: cfg.MaxHeaderBytes,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
	}
	s.register(svc)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			var err error
			if e, ok := x.(error); ok {
				err = e
			}

			s.logger.Log("error", err, "message", "the server encountered a panic")

			w.WriteHeader(http.StatusInternalServerError)
			_, werr := w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			if werr != nil {
				s.logger.Log("error", werr, "message", "unable to respond post-panic")
			}

			// if we have an error client, send out a report
			if s.errs == nil {
				return
			}

			s.errs.Report(errorreporting.Entry{
				Req:   r,
				Error: err,
			})
			s.errs.Flush()
		}
	}()

	s.svc.HTTPMiddleware(s.mux).ServeHTTP(w, r)
}

func (s *Server) register(svc Service) {
	s.svc = svc
	opts := []httptransport.ServerOption{
		// populate context with helpful keys
		httptransport.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
			ctx = httptransport.PopulateRequestContext(ctx, r)
			// add google trace header to use in tracing and logging
			return context.WithValue(ctx, ContextKeyCloudTraceContext,
				r.Header.Get("X-Cloud-Trace-Context"))
		}),
		// inject the server logger into every request context
		httptransport.ServerBefore(func(ctx context.Context, _ *http.Request) context.Context {
			return context.WithValue(ctx, logKey, AddLogKeyVals(ctx, s.logger))
		}),
	}
	opts = append(opts, svc.HTTPOptions()...)

	const warmupPath = "/_ah/warmup"
	var (
		healthzFound bool
		warmupFound  bool
	)
	// register all endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range svc.HTTPEndpoints() {
		for method, ep := range epMethods {

			// check if folks are supplying their own healthcheck
			if method == http.MethodGet && path == s.cfg.HealthCheckPath {
				healthzFound = true
			}

			// check for a GAE "warm up" request endpoint
			if method == http.MethodGet && path == warmupPath {
				warmupFound = true
			}

			// just pass the http.Request in if no decoder provided
			if ep.Decoder == nil {
				ep.Decoder = basicDecoder
			}
			// default to the httptransport helper
			if ep.Encoder == nil {
				ep.Encoder = httptransport.EncodeJSONResponse
			}
			s.mux.Handle(method, path,
				httptransport.NewServer(
					svc.Middleware(ep.Endpoint),
					ep.Decoder,
					ep.Encoder,
					append(opts, ep.Options...)...))
		}
	}

	// register a simple health check if none provided
	if !healthzFound {
		s.mux.Handle(http.MethodGet, s.cfg.HealthCheckPath,
			httptransport.NewServer(
				svc.Middleware(okEndpoint),
				basicDecoder,
				httptransport.EncodeJSONResponse,
				opts...))
	}

	// register a warmup request for App Engine apps that dont have one already.
	if !warmupFound {
		s.mux.Handle(http.MethodGet, warmupPath,
			httptransport.NewServer(
				svc.Middleware(okEndpoint),
				basicDecoder,
				httptransport.EncodeJSONResponse,
				opts...))
	}

	// add all pprof endpoints by default to HTTP
	registerPprof(s.cfg, s.mux)

	gdesc := svc.RPCServiceDesc()
	if gdesc == nil {
		return
	}

	inters := []grpc.UnaryServerInterceptor{
		grpc.UnaryServerInterceptor(
			// inject logger into gRPC server and hook in go-kit middleware
			func(ctx ocontext.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctx = context.WithValue(ctx, logKey, AddLogKeyVals(ctx, s.logger))
				return svc.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
					return handler(ctx, req)
				})(ctx, req)
			},
		),
	}
	if mw := svc.RPCMiddleware(); mw != nil {
		inters = append(inters, mw)
	}

	s.gsvr = grpc.NewServer(append(svc.RPCOptions(),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(inters...)))...)

	s.gsvr.RegisterService(gdesc, svc)
}

func okEndpoint(ctx context.Context, _ interface{}) (interface{}, error) {
	return "OK", nil
}

func basicDecoder(_ context.Context, r *http.Request) (interface{}, error) {
	return r, nil
}

func (s *Server) start() error {
	go func() {
		err := s.svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Log(
				"error", err,
				"message", "HTTP server error - initiating shutting down")
			s.stop()
		}
	}()

	s.logger.Log("message",
		fmt.Sprintf("listening on HTTP port: %d", s.cfg.HTTPPort))

	if s.gsvr != nil {
		gaddr := fmt.Sprintf(":%d", s.cfg.RPCPort)
		lis, err := net.Listen("tcp", gaddr)
		if err != nil {
			return errors.Wrap(err, "failed to listen to RPC port")
		}

		go func() {
			err := s.gsvr.Serve(lis)
			// the gRPC server _always_ returns non-nil
			// this filters out the known err we don't care about logging
			if (err != nil) && strings.Contains(err.Error(), "use of closed network connection") {
				err = nil
			}
			if err != nil {
				s.logger.Log(
					"error", err,
					"message", "gRPC server error - initiating shutting down")
				s.stop()
			}
		}()
		s.logger.Log("message",
			fmt.Sprintf("listening on RPC port: %d", s.cfg.RPCPort))
	}

	go func() {
		exit := <-s.exit

		// stop the listener with timeout
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		defer func() {
			// flush the logger after server shuts down
			if s.logClose != nil {
				s.logClose()
			}

			if s.errs != nil {
				s.errs.Close()
			}
		}()

		if shutdown, ok := s.svc.(Shutdowner); ok {
			shutdown.Shutdown()
		}
		if s.gsvr != nil {
			s.gsvr.GracefulStop()
		}
		exit <- s.svr.Shutdown(ctx)
	}()

	return nil
}

func (s *Server) stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}

func registerPprof(cfg Config, mx Router) {
	if !cfg.EnablePProf {
		return
	}
	mx.HandleFunc(http.MethodGet, "/debug/pprof/", pprof.Index)
	mx.HandleFunc(http.MethodGet, "/debug/pprof/cmdline", pprof.Cmdline)
	mx.HandleFunc(http.MethodGet, "/debug/pprof/profile", pprof.Profile)
	mx.HandleFunc(http.MethodGet, "/debug/pprof/symbol", pprof.Symbol)
	mx.HandleFunc(http.MethodGet, "/debug/pprof/trace", pprof.Trace)
	// Manually add support for paths linked to by index page at /debug/pprof/
	mx.Handle(http.MethodGet, "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mx.Handle(http.MethodGet, "/debug/pprof/heap", pprof.Handler("heap"))
	mx.Handle(http.MethodGet, "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mx.Handle(http.MethodGet, "/debug/pprof/block", pprof.Handler("block"))
}
