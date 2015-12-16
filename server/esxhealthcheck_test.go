package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestESXHealthCheckDeployer(t *testing.T) {
	// changing so the tests don't take forever, but long enough so tests fail
	ESXLoadBalancerNotReadyDuration = 10 * time.Second

	hc := NewESXHealthCheck()
	am := NewActivityMonitor()
	hc.Start(am)

	// setup an LB-like request with an easy IP header
	lbIP := "1.1.1.1"
	lbReq, _ := http.NewRequest("GET", hc.Path()+"?deployer=true", nil)
	lbReq.Header.Add("X-Real-IP", lbIP)

	wr := httptest.NewRecorder()
	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusOK {
		t.Errorf("ESXHealthCheck expected 200 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); !strings.HasPrefix(gotBody, "ok-") {
		t.Errorf("ESXHealthCheck expected response body to start with 'ok-', got %s", gotBody)
	}

	// gorountine to make sure Stop actually stops on time
	done := make(chan bool)
	go func() {
		stopped := make(chan bool)
		go func() {
			hc.Stop()
			stopped <- true
		}()
		select {
		case <-stopped:
		case <-time.After(2 * time.Second):
			t.Errorf("ESXHealthCheck.Stop() was still stopping after 2 seconds with no LBs checking!")
		}
		done <- true
	}()
	// give it a moment
	time.Sleep(20 * time.Millisecond)

	wr = httptest.NewRecorder()

	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusServiceUnavailable {
		t.Errorf("ESXHealthCheck expected 503 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); gotBody != "service unavailable\n" {
		t.Errorf("ESXHealthCheck expected response body to start with 'service unavailable', got %s", gotBody)
	}

	<-done
}

func TestESXHealthCheckLB(t *testing.T) {
	hc := NewESXHealthCheck()
	am := NewActivityMonitor()
	// changing so the tests don't take forever
	ESXLoadBalancerNotReadyDuration = 1 * time.Second

	hc.Start(am)

	// setup an LB-like request with an easy IP header
	lbIP := "1.1.1.1"
	lbReq, _ := http.NewRequest("GET", hc.Path(), nil)
	lbReq.Header.Add("X-Real-IP", lbIP)

	wr := httptest.NewRecorder()
	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusOK {
		t.Errorf("ESXHealthCheck expected 200 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); !strings.HasPrefix(gotBody, "ok-") {
		t.Errorf("ESXHealthCheck expected response body to start with 'ok-', got %s", gotBody)
	}

	// gorountine to make sure Stop actually stops on time
	done := make(chan bool)
	go func() {
		stopped := make(chan bool)
		go func() {
			hc.Stop()
			stopped <- true
		}()
		select {
		case <-stopped:
		case <-time.After(ESXLoadBalancerNotReadyDuration * (2 * time.Second)):
			t.Errorf("ESXHealthCheck.Stop() was still stopping 2x after ESXLoadBalancerNotReadyDuration")
		}
		done <- true
	}()
	// give it a moment
	time.Sleep(20 * time.Millisecond)

	// let the LB see a bad request
	wr = httptest.NewRecorder()
	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusServiceUnavailable {
		t.Errorf("ESXHealthCheck expected 503 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); gotBody != "service unavailable\n" {
		t.Errorf("ESXHealthCheck expected response body to start with 'service unavailable', got %s", gotBody)
	}

	// wait for the healthcheck to stop
	<-done
}

func TestESXHealthCheckActiveRequests(t *testing.T) {
	// changing so the tests don't take forever, but long enough so tests fail
	ESXLoadBalancerNotReadyDuration = 5 * time.Second
	ESXShutdownTimeout = 5 * time.Second

	hc := NewESXHealthCheck()
	am := NewActivityMonitor()
	hc.Start(am)

	// setup an LB-like request with an easy IP header
	lbIP := "1.1.1.1"
	lbReq, _ := http.NewRequest("GET", hc.Path(), nil)
	lbReq.Header.Add("X-Real-IP", lbIP)

	wr := httptest.NewRecorder()
	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusOK {
		t.Errorf("ESXHealthCheck expected 200 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); !strings.HasPrefix(gotBody, "ok-") {
		t.Errorf("ESXHealthCheck expected response body to start with 'ok-', got %s", gotBody)
	}

	// WHOA, AN ACTIVE REQUEST ENTERS THE SCENE FROM STAGE LEFT!
	am.CountRequest()

	// gorountine to make sure Stop actually stops on time
	done := make(chan bool)
	go func() {
		stopped := make(chan bool)
		go func() {
			err := hc.Stop()
			if err == nil {
				t.Error("ESXHealthCheck expected a shutdown timeout error because there was an active request")
			}
			stopped <- true
		}()
		select {
		case <-stopped:
		case <-time.After(ESXShutdownTimeout * 2):
			t.Errorf("ESXHealthCheck.Stop() was still stopping after 2 seconds with no LBs checking!")
		}
		done <- true
	}()
	// give it a moment
	time.Sleep(20 * time.Millisecond)

	wr = httptest.NewRecorder()

	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusServiceUnavailable {
		t.Errorf("ESXHealthCheck expected 503 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); gotBody != "service unavailable\n" {
		t.Errorf("ESXHealthCheck expected response body to start with 'service unavailable', got %s", gotBody)
	}

	<-done
}
func TestESXHealthCheckBadIP(t *testing.T) {
	// changing so the tests don't take forever, but long enough so tests fail
	ESXLoadBalancerNotReadyDuration = 10 * time.Second

	hc := NewESXHealthCheck()
	am := NewActivityMonitor()
	hc.Start(am)

	// setup an LB-like request with an easy IP header
	lbReq, _ := http.NewRequest("GET", hc.Path(), nil)

	wr := httptest.NewRecorder()
	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusOK {
		t.Errorf("ESXHealthCheck expected 200 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); !strings.HasPrefix(gotBody, "ok-") {
		t.Errorf("ESXHealthCheck expected response body to start with 'ok-', got %s", gotBody)
	}

	// gorountine to make sure Stop actually stops on time
	done := make(chan bool)
	go func() {
		stopped := make(chan bool)
		go func() {
			hc.Stop()
			stopped <- true
		}()
		select {
		case <-stopped:
		case <-time.After(2 * time.Second):
			t.Errorf("ESXHealthCheck.Stop() was still stopping after 2 seconds with no LBs checking!")
		}
		done <- true
	}()
	// give it a moment
	time.Sleep(20 * time.Millisecond)

	wr = httptest.NewRecorder()

	hc.ServeHTTP(wr, lbReq)

	if wr.Code != http.StatusServiceUnavailable {
		t.Errorf("ESXHealthCheck expected 503 response code, got %d", wr.Code)
	}

	if gotBody := wr.Body.String(); gotBody != "service unavailable\n" {
		t.Errorf("ESXHealthCheck expected response body to start with 'service unavailable', got %s", gotBody)
	}

	<-done
}
