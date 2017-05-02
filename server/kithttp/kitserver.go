package kithttp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	gserver "github.com/NYTimes/gizmo/server"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

type server struct {
	Logger log.Logger

	mux Router

	cfg gserver.Config

	// exit chan for graceful shutdown
	exit chan chan error
}

// newServer will init the mux and register all endpoints.
func newServer(cfg gserver.Config, opts ...RouterOption) server {
	if len(opts) == 0 {
		// select the default router
		opts = append(opts, RouterSelect(""))
	}
	var r Router
	for _, opt := range opts {
		r = opt(r)
	}
	return server{
		cfg:  cfg,
		mux:  r,
		exit: make(chan chan error),
	}
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hand the request off to the router
	s.mux.ServeHTTP(w, r)
}

// Register will accept and register server, JSONService or MixedService implementations.
func (s server) Register(svc Service) error {
	var (
		jseps map[string]map[string]HTTPEndpoint
		peps  map[string]map[string]HTTPEndpoint
	)
	switch svc.(type) {
	case MixedService:
		jseps = svc.(JSONService).JSONEndpoints()
		peps = svc.(ProtoService).ProtoEndpoints()
	case JSONService:
		jseps = svc.(JSONService).JSONEndpoints()
	case ProtoService:
		peps = svc.(ProtoService).ProtoEndpoints()
	default:
		return errors.New("services for servers must implement one of the Service interface extensions")
	}

	opts := defaultOpts
	opts = append(opts, svc.Options()...)

	// register all JSON endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range jseps {
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

	// register all Protobuf endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range peps {
		for method, ep := range epMethods {
			// just pass the http.Request in if no decoder provided
			if ep.Decoder == nil {
				ep.Decoder = func(_ context.Context, r *http.Request) (interface{}, error) {
					return r, nil
				}
			}
			// default to the a protobuf helper
			if ep.Encoder == nil {
				ep.Encoder = EncodeProtoResponse
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

func (s server) Start() error {
	// TODO(jprobinson) add basic default health check

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
			s.Logger.Log("server error", err)
			panic("server error" + err.Error())
		}
	}()
	s.Logger.Log(fmt.Sprintf("Listening on %s", addr))

	go func() {
		exit := <-s.exit

		// stop the listener with timeout (5mins for now until we abstract to cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		exit <- srv.Shutdown(ctx)
	}()
	return nil
}

func (s server) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
