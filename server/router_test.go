package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGorillaRoute(t *testing.T) {
	cfg := &Config{HealthCheckType: "simple", HealthCheckPath: "/status"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkSimpleService{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/svc/v1/1/blah/:something", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	srvr.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("SimpleHealthCheck expected 200 response code, got %d", w.Code)
	}

	wantBody := "blah"
	if gotBody := w.Body.String(); gotBody != wantBody {
		t.Errorf("Gorilla route expected response body to be %q, got %q", wantBody, gotBody)
	}
}

func TestStdlibRoute(t *testing.T) {
	cfg := &Config{RouterType: "stdlib"}
	srvr := NewSimpleServer(cfg)
	RegisterHealthHandler(cfg, srvr.monitor, srvr.mux)
	srvr.Register(&benchmarkSimpleService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/svc/v1/2", nil)
	r.RemoteAddr = "0.0.0.0:8080"

	srvr.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("stdlib route expected 200 response code, got %d", w.Code)
	}

	wantBody := "ok"
	if gotBody := w.Body.String(); gotBody != wantBody {
		t.Errorf("stdlib route expected response body to be %q, got %q", wantBody, gotBody)
	}
}
