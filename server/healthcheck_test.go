package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSimpleHealthCheck(t *testing.T) {
	path := "/status.txt"
	s := NewSimpleHealthCheck(path)
	a := NewActivityMonitor()
	s.Start(a)

	req, _ := http.NewRequest("GET", path, nil)
	wr := httptest.NewRecorder()

	s.ServeHTTP(wr, req)

	if wr.Code != http.StatusOK {
		t.Errorf("SimpleHealthCheck expected 200 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); !strings.HasPrefix(gotBody, "ok-") {
		t.Errorf("SimpleHealthCheck expected response body to start with 'ok-', got %s", gotBody)
	}

	s.Stop()

	wr = httptest.NewRecorder()

	s.ServeHTTP(wr, req)

	if wr.Code != http.StatusOK {
		t.Errorf("SimpleHealthCheck expected 200 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); !strings.HasPrefix(gotBody, "ok-") {
		t.Errorf("SimpleHealthCheck expected response body to start with 'ok-', got %s", gotBody)
	}

}
