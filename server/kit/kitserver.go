package kit

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	ocontext "golang.org/x/net/context"
	"google.golang.org/grpc"

	gserver "github.com/NYTimes/gizmo/server"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

type server struct {
	logger log.Logger

	mux Router

	cfg gserver.Config

	svc Service

	// exit chan for graceful shutdown
	exit chan chan error
}

// newServer will init the mux and register all endpoints.
func newServer(cfg gserver.Config, opts ...RouterOption) *server {
	if len(opts) == 0 {
		// select the default router
		opts = append(opts, RouterSelect(""))
	}
	var r Router
	for _, opt := range opts {
		r = opt(r)
	}
	return &server{
		cfg:    cfg,
		mux:    r,
		exit:   make(chan chan error),
		logger: log.NewJSONLogger(log.NewSyncWriter(os.Stdout)),
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hand the request off to the router
	s.mux.ServeHTTP(w, r)
}

// Register will accept and register server, JSONService or MixedService implementations.
func (s *server) Register(svc Service) error {
	s.svc = svc
	opts := []httptransport.ServerOption{
		// inject the server logger into every request context
		httptransport.ServerBefore(func(ctx context.Context, _ *http.Request) context.Context {
			return context.WithValue(ctx, logKey, s.logger)
		}),
	}
	opts = append(opts, defaultHTTPOpts...)
	opts = append(opts, svc.HTTPOptions()...)

	// register all endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range svc.HTTPEndpoints() {
		for method, ep := range epMethods {
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

	return nil
}

func (s *server) Start() error {
	// TODO(jprobinson) add basic default health check

	// HTTP SERVER!
	addr := fmt.Sprintf(":%d", s.cfg.HTTPPort)
	srv := http.Server{
		Handler:        s,
		MaxHeaderBytes: maxHeaderBytes,
		Addr:           addr,
	}

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
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			s.logger.Log("server error", err)
			panic("server error" + err.Error())
		}
	}()
	s.logger.Log("listening on HTTP port", addr)

	// gRPC SERVER!
	var gsrv *grpc.Server
	if desc := s.svc.RPCServiceDesc(); desc != nil {
		gaddr := fmt.Sprintf(":%d", s.cfg.RPCPort)
		lis, err := net.Listen("tcp", gaddr)
		if err != nil {
			panic("failed to listen: " + err.Error())
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
		gopts = append(gopts, s.svc.RPCOptions()...)
		gsrv = grpc.NewServer(gopts...)
		gsrv.RegisterService(desc, s.svc)
		go func() {
			err := gsrv.Serve(lis)
			if err != nil {
				s.logger.Log("gRPC server error", err)
				panic("gRPC server error" + err.Error())
			}
		}()
		s.logger.Log("listening on RPC port", gaddr)
	}

	go func() {
		exit := <-s.exit

		// stop the listener with timeout (5mins for now until we abstract to cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if gsrv != nil {
			gsrv.GracefulStop()
		}
		exit <- srv.Shutdown(ctx)
	}()
	return nil
}

func (s *server) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
