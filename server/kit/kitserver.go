package kit

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	ocontext "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
)

type server struct {
	logger log.Logger

	mux Router

	cfg Config

	svr  *http.Server
	gsvr *grpc.Server

	// exit chan for graceful shutdown
	exit chan chan error
}

func newServer(svc Service) *server {
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

	s := &server{
		cfg:    cfg,
		mux:    r,
		exit:   make(chan chan error),
		logger: log.NewJSONLogger(log.NewSyncWriter(os.Stdout)),
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

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hand the request off to the router
	s.mux.ServeHTTP(w, r)
}

func (s *server) register(svc Service) {
	opts := []httptransport.ServerOption{
		// populate context with helpful keys
		httptransport.ServerBefore(
			httptransport.PopulateRequestContext,
		),
		// inject the server logger into every request context
		httptransport.ServerBefore(func(ctx context.Context, _ *http.Request) context.Context {
			return context.WithValue(ctx, logKey, s.logger)
		}),
	}
	opts = append(opts, svc.HTTPOptions()...)

	var healthzFound bool
	// register all endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range svc.HTTPEndpoints() {
		for method, ep := range epMethods {

			// check if folks are supplying their own healthcheck
			if method == http.MethodGet && path == s.cfg.HealthCheckPath {
				healthzFound = true
			}

			// just pass the http.Request in if no decoder provided
			if ep.Decoder == nil {
				ep.Decoder = func(_ context.Context, r *http.Request) (interface{}, error) {
					return r, nil
				}
			}
			// default to the httptransport helper
			if ep.Encoder == nil {
				ep.Encoder = httptransport.EncodeJSONResponse
			}
			s.mux.Handle(method, path, svc.HTTPMiddleware(
				httptransport.NewServer(
					svc.Middleware(ep.Endpoint),
					ep.Decoder,
					ep.Encoder,
					append(opts, ep.Options...)...)))
		}
	}

	// register a simple health check if none provided
	if !healthzFound {
		s.mux.HandleFunc(http.MethodGet, s.cfg.HealthCheckPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK")
		})
	}

	gdesc := svc.RPCServiceDesc()
	if gdesc == nil {
		return
	}

	gopts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc.UnaryServerInterceptor(
					// inject logger into gRPC server
					func(ctx ocontext.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
						ctx = context.WithValue(ctx, logKey, s.logger)
						return handler(ctx, req)
					}),
			),
		),
	}
	gopts = append(gopts, svc.RPCOptions()...)
	s.gsvr = grpc.NewServer(gopts...)
	s.gsvr.RegisterService(gdesc, svc)
}

func (s *server) start() error {
	go func() {
		err := s.svr.ListenAndServe()
		if err != nil {
			s.logger.Log("server error", err,
				"initiating shutting down", true)
			s.stop()
		}
	}()
	s.logger.Log("listening on HTTP port", s.cfg.HTTPPort)

	if s.gsvr != nil {
		gaddr := fmt.Sprintf(":%d", s.cfg.RPCPort)
		lis, err := net.Listen("tcp", gaddr)
		if err != nil {
			return errors.Wrap(err, "failed to listen to RPC port")
		}

		go func() {
			err := s.gsvr.Serve(lis)
			if err != nil {
				s.logger.Log("gRPC server error", err,
					"initiating shutting down", true)
				s.stop()
			}
		}()
		s.logger.Log("listening on RPC port", s.cfg.RPCPort)
	}

	go func() {
		exit := <-s.exit

		// stop the listener with timeout
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if s.gsvr != nil {
			s.gsvr.GracefulStop()
		}
		exit <- s.svr.Shutdown(ctx)
	}()

	return nil
}

func (s *server) stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
