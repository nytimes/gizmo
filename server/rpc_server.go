package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/NYTimes/gizmo/config"

	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// RPCServer is an experimental server that serves a gRPC server on one
// port and the same endpoints via JSON on another port.
type RPCServer struct {
	cfg *config.Server

	// exit chan for graceful shutdown
	exit chan chan error

	// server for handling RPC requests
	srvr *grpc.Server

	// mux for routing HTTP requests
	mux *mux.Router

	// tracks active requests
	monitor *ActivityMonitor

	// registry for collecting metrics
	registry metrics.Registry
}

// NewRPCServer will instantiate a new experimental RPCServer with the given config.
func NewRPCServer(cfg *config.Server) *RPCServer {
	if cfg == nil {
		cfg = &config.Server{}
	}
	mx := mux.NewRouter()
	if cfg.NotFoundHandler != nil {
		mx.NotFoundHandler = cfg.NotFoundHandler
	}
	return &RPCServer{
		cfg:      cfg,
		srvr:     grpc.NewServer(),
		mux:      mx,
		exit:     make(chan chan error),
		monitor:  NewActivityMonitor(),
		registry: metrics.NewRegistry(),
	}
}

// Register will attempt to register the given RPCService with the server.
// If any other types are passed, Register will panic.
func (r *RPCServer) Register(svc Service) error {
	rpcsvc, ok := svc.(RPCService)
	if !ok {
	}

	// register RPC
	desc, grpcSvc := rpcsvc.Service()
	r.srvr.RegisterService(desc, grpcSvc)
	// register endpoints
	for _, mthd := range desc.Methods {
		registerRPCMetrics(mthd.MethodName, r.registry)
	}

	// register HTTP
	prefix := rpcsvc.Prefix()
	sr := r.mux.PathPrefix(prefix).Subrouter()

	// loop through json endpoints and register them
	for path, epMethods := range rpcsvc.JSONEndpoints() {
		for method, ep := range epMethods {
			endpointName := metricName(prefix, path, method)
			// set the function handle and register is to metrics
			sr.Handle(path, Timed(CountedByStatusXX(
				rpcsvc.Middleware(JSONToHTTP(rpcsvc.JSONMiddleware(ep))),
				endpointName+".STATUS-COUNT", r.registry),
				endpointName+".DURATION", r.registry),
			).Methods(method)
		}
	}

	RegisterProfiler(r.cfg, r.mux)

	return nil
}

// Start start the RPC server.
func (r *RPCServer) Start() error {

	StartServerMetrics(r.cfg, r.registry)

	rl, err := net.Listen("tcp", fmt.Sprintf(":%d", r.cfg.RPCPort))
	if err != nil {
		return err
	}

	func() {
		if err := r.srvr.Serve(rl); err != nil {
			panic("unable to serve: " + err.Error())
		}
	}()

	// setup HTTP
	healthHandler := RegisterHealthHandler(r.cfg, r.monitor, r.mux)
	r.cfg.HealthCheckPath = healthHandler.Path()
	srv := http.Server{
		Handler:        r,
		MaxHeaderBytes: maxHeaderBytes,
	}
	var hl net.Listener
	hl, err = net.Listen("tcp", fmt.Sprintf(":%d", r.cfg.HTTPPort))
	if err != nil {
		return err
	}

	func() {
		if err := srv.Serve(hl); err != nil {
			panic("unable to serve: " + err.Error())
		}
	}()

	// join the LB
	go func() {
		exit := <-r.exit

		if err := healthHandler.Stop(); err != nil {
		}

		r.srvr.Stop()
		exit <- hl.Close()
	}()

	return nil
}

// Stop will signal the RPC server to stop and block until it does.
func (r *RPCServer) Stop() error {
	ch := make(chan error)
	r.exit <- ch
	return <-ch
}

// ServeHTTP is RPCServer's hook for metrics and safely executing each request.
func (r *RPCServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	AddIPToContext(req)

	// only count non-LB requests
	if req.URL.Path != r.cfg.HealthCheckPath {
		r.monitor.CountRequest()
		defer r.monitor.UncountRequest()
	}

	r.safelyExecuteHTTPRequest(w, req)
}

// executeRequestSafely will prevent a panic in a request from bringing the server down.
func (r *RPCServer) safelyExecuteHTTPRequest(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			// register a panic'd request with our metrics
			errCntr := metrics.GetOrRegisterCounter("PANIC", r.registry)
			errCntr.Inc(1)

			// give the users our deepest regrets
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write(UnexpectedServerError); err != nil {
			}

		}
	}()

	// hand the request off to gorilla
	r.mux.ServeHTTP(w, req)
}

// MonitorRPCRequest should be deferred by any RPC method that would like to have
// metrics and access logging, participate in graceful shutdowns and safely recover from panics.
func MonitorRPCRequest() func(ctx context.Context, methodName string, err error) {
	start := time.Now()
	return func(ctx context.Context, methodName string, err error) {
		if x := recover(); x != nil {
			// register a panic'd request with our metrics
			errCntr := metrics.GetOrRegisterCounter("RPC PANIC", metrics.DefaultRegistry)
			errCntr.Inc(1)

			// give the users our deepest regrets
			err = errors.New(string(UnexpectedServerError))
		}

		m := rpcEndpointMetrics["rpc."+methodName]
		if err == nil {
			m.SuccessCounter.Inc(1)
		} else {
			m.ErrorCounter.Inc(1)
		}
		m.Timer.UpdateSince(start)

	}
}

var rpcEndpointMetrics = map[string]*rpcMetrics{}

type rpcMetrics struct {
	Timer          metrics.Timer
	SuccessCounter metrics.Counter
	ErrorCounter   metrics.Counter
}

func registerRPCMetrics(name string, registry metrics.Registry) {
	name = "rpc." + name
	m := &rpcMetrics{}
	m.Timer = metrics.NewTimer()
	if err := registry.Register(name+".DURATION", m.Timer); nil != err {
		return
	}

	m.SuccessCounter = metrics.NewCounter()
	if err := registry.Register(name+".ERROR", m.SuccessCounter); nil != err {
		return
	}

	m.ErrorCounter = metrics.NewCounter()
	if err := registry.Register(name+".SUCCESS", m.ErrorCounter); nil != err {
		return
	}

	rpcEndpointMetrics[name] = m
}
