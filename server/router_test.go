package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFastRoute(t *testing.T) {
	cfg := &Config{RouterType: "fast", HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkSimpleService{true})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/svc/v1/1/{something}/blah", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	srvr.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("SimpleHealthCheck expected 200 response code, got %d", w.Code)
	}

	wantBody := "blah"
	if gotBody := w.Body.String(); gotBody != wantBody {
		t.Errorf("Fast route expected response body to be %q, got %q", wantBody, gotBody)
	}
}

func TestGorillaRoute(t *testing.T) {
	cfg := &Config{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkSimpleService{false})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/svc/v1/1/blah/:something", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	srvr.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("SimpleHealthCheck expected 200 response code, got %d", w.Code)
	}

	wantBody := "blah"
	if gotBody := w.Body.String(); gotBody != wantBody {
		t.Errorf("Fast route expected response body to be %q, got %q", wantBody, gotBody)
	}
}
