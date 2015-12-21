package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NYTimes/gizmo/config"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
)

func BenchmarkSimpleServer_NoParam(b *testing.B) {
	cfg := &config.Server{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkSimpleService{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/svc/v1/2", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	for i := 0; i < b.N; i++ {
		srvr.ServeHTTP(w, r)
	}

	metrics.DefaultRegistry.UnregisterAll()
}

func BenchmarkSimpleServer_WithParam(b *testing.B) {
	cfg := &config.Server{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkSimpleService{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/svc/v1/1/blah", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	for i := 0; i < b.N; i++ {
		srvr.ServeHTTP(w, r)
	}

	metrics.DefaultRegistry.UnregisterAll()
}

type benchmarkSimpleService struct{}

func (s *benchmarkSimpleService) Prefix() string {
	return "/svc/v1"
}

func (s *benchmarkSimpleService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/1/{something}": map[string]http.HandlerFunc{
			"GET": s.GetSimple,
		},
		"/2": map[string]http.HandlerFunc{
			"GET": s.GetSimpleNoParam,
		},
	}
}

func (s *benchmarkSimpleService) Middleware(h http.Handler) http.Handler {
	return h
}

func (s *benchmarkSimpleService) GetSimple(w http.ResponseWriter, r *http.Request) {
	something := mux.Vars(r)["something"]
	fmt.Fprint(w, something)
}

func (s *benchmarkSimpleService) GetSimpleNoParam(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func BenchmarkJSONServer_JSONPayload(b *testing.B) {
	cfg := &config.Server{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkJSONService{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/svc/v1/1", bytes.NewBufferString(`{"hello":"hi","howdy":"yo"}`))
	r.RemoteAddr = "0.0.0.0:8080"

	for i := 0; i < b.N; i++ {
		srvr.ServeHTTP(w, r)
	}

	metrics.DefaultRegistry.UnregisterAll()
}
func BenchmarkJSONServer_NoParam(b *testing.B) {
	cfg := &config.Server{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkJSONService{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/svc/v1/2", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	for i := 0; i < b.N; i++ {
		srvr.ServeHTTP(w, r)
	}

	metrics.DefaultRegistry.UnregisterAll()
}
func BenchmarkJSONServer_WithParam(b *testing.B) {
	cfg := &config.Server{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkJSONService{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/svc/v1/3/blah", bytes.NewBufferString(`{"hello":"hi","howdy":"yo"}`))
	r.RemoteAddr = "0.0.0.0:8080"

	for i := 0; i < b.N; i++ {
		srvr.ServeHTTP(w, r)
	}

	metrics.DefaultRegistry.UnregisterAll()
}

type benchmarkJSONService struct{}

func (s *benchmarkJSONService) Prefix() string {
	return "/svc/v1"
}

func (s *benchmarkJSONService) JSONEndpoints() map[string]map[string]JSONEndpoint {
	return map[string]map[string]JSONEndpoint{
		"/1": map[string]JSONEndpoint{
			"PUT": s.PutJSON,
		},
		"/2": map[string]JSONEndpoint{
			"GET": s.GetJSON,
		},
		"/3/{something}": map[string]JSONEndpoint{
			"GET": s.GetJSONParam,
		},
	}
}

func (s *benchmarkJSONService) JSONMiddleware(e JSONEndpoint) JSONEndpoint {
	return e
}

func (s *benchmarkJSONService) Middleware(h http.Handler) http.Handler {
	return h
}

func (s *benchmarkJSONService) PutJSON(r *http.Request) (int, interface{}, error) {
	var hello testJSON
	err := json.NewDecoder(r.Body).Decode(&hello)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	return http.StatusOK, hello, nil
}

func (s *benchmarkJSONService) GetJSON(r *http.Request) (int, interface{}, error) {
	return http.StatusOK, &testJSON{"hi", "howdy"}, nil
}

func (s *benchmarkJSONService) GetJSONParam(r *http.Request) (int, interface{}, error) {
	something := mux.Vars(r)["something"]
	return http.StatusOK, &testJSON{"hi", something}, nil
}

type testJSON struct {
	Hello string `json:"hello"`
	Howdy string `json:"howdy"`
}

type testMixedService struct{}

func (s *testMixedService) Prefix() string {
	return "/svc/v1"
}

func (s *testMixedService) JSONEndpoints() map[string]map[string]JSONEndpoint {
	return map[string]map[string]JSONEndpoint{
		"/json": map[string]JSONEndpoint{
			"GET": s.GetJSON,
		},
	}
}

func (s *testMixedService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/simple": map[string]http.HandlerFunc{
			"GET": s.GetSimple,
		},
	}
}

func (s *testMixedService) GetSimple(w http.ResponseWriter, r *http.Request) {
	something := mux.Vars(r)["something"]
	fmt.Fprint(w, something)
}

func (s *testMixedService) GetJSON(r *http.Request) (int, interface{}, error) {
	return http.StatusOK, &testJSON{"hi", "howdy"}, nil
}

func (s *testMixedService) JSONMiddleware(e JSONEndpoint) JSONEndpoint {
	return e
}

func (s *testMixedService) Middleware(h http.Handler) http.Handler {
	return h
}

type testInvalidService struct{}

func (s *testInvalidService) Prefix() string {
	return "/svc/v1"
}

func (s *testInvalidService) Middleware(h http.Handler) http.Handler {
	return h
}

func TestFactory(*testing.T) {
	// with config:
	cfg := &config.Server{HealthCheckType: "simple", HealthCheckPath: "/status"}
	NewSimpleServer(cfg)

	// without config:
	NewSimpleServer(nil)

}

func TestBasicRegistration(t *testing.T) {
	s := NewSimpleServer(nil)
	services := []Service{
		&benchmarkSimpleService{},
		&benchmarkJSONService{},
		&testMixedService{},
	}
	for _, svc := range services {
		if err := s.Register(svc); err != nil {
			t.Errorf("Basic registration of services should not encounter an error: %s\n", err)
		}
		// need to unregister DefaultRegistry in between service registrations
		metrics.DefaultRegistry.UnregisterAll()
	}

	if err := s.Register(&testInvalidService{}); err == nil {
		t.Error("Invalid services should produce an error in service registration")
	}
}
