package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
)

func TestCounterByStatusXX(t *testing.T) {
	tests := []int{111, 222, 333, 444, 555}
	statuses := make(chan int, 1)
	provider := newMockProvider()

	counter := CountedByStatusXX(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := <-statuses
		w.WriteHeader(status)
		if bod, _ := ioutil.ReadAll(r.Body); string(bod) != "blah" {
			t.Errorf("CountedByStatusXX expected the request body to be 'blah', got '%s'", string(bod))
		}
		r.Body.Close()
	}), "counted", provider)

	for _, given := range tests {
		statuses <- given
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "http://yup.com/foo", bytes.NewBufferString("blah"))
		counter.ServeHTTP(w, r)
		if given != w.Code {
			t.Errorf("CountedByStatusXX expected response code of %d, got %d", given, w.Code)
		}
	}

	close(statuses)

	if cnt := provider.counters["counted-1xx"].lastAdd; cnt != 1 {
		t.Errorf("CountedByStatusXX expected 1xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := provider.counters["counted-2xx"].lastAdd; cnt != 1 {
		t.Errorf("CountedByStatusXX expected 2xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := provider.counters["counted-3xx"].lastAdd; cnt != 1 {
		t.Errorf("CountedByStatusXX expected 3xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := provider.counters["counted-4xx"].lastAdd; cnt != 1 {
		t.Errorf("CountedByStatusXX expected 4xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := provider.counters["counted-5xx"].lastAdd; cnt != 1 {
		t.Errorf("CountedByStatusXX expected 5xx counter to have a count of 1, got %d", cnt)
	}
}

func TestTimer(t *testing.T) {
	provider := newMockProvider()

	r, _ := http.NewRequest("POST", "http://uhhuh.io/", bytes.NewBufferString("yerp"))
	timer := Timed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		if bod, _ := ioutil.ReadAll(r.Body); string(bod) != "yerp" {
			t.Errorf("Timer expected the request body to be 'yerp', got '%s'", string(bod))
		}
		r.Body.Close()
	}), "timed", provider)
	w := httptest.NewRecorder()
	timer.ServeHTTP(w, r)

	if dur := provider.histograms["timed"].lastObserved; dur < 200 || dur > 300 {
		t.Errorf("Timer expected Max() to return between 200 and 300 ms, got %d", dur)
	}
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		counters:   map[string]*mockCounter{},
		histograms: map[string]*mockHistogram{},
		gauges:     map[string]*mockGauge{},
	}
}

type mockProvider struct {
	mtx        sync.Mutex
	counters   map[string]*mockCounter
	gauges     map[string]*mockGauge
	histograms map[string]*mockHistogram
}

func (p *mockProvider) NewCounter(name string, help string) metrics.Counter {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	c := &mockCounter{}
	p.counters[name] = c
	return c
}

func (p *mockProvider) NewHistogram(name string, help string, min int64, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	h := &mockHistogram{}
	p.histograms[name] = h
	return h, nil
}

func (p *mockProvider) NewGauge(name string, help string) metrics.Gauge {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	g := &mockGauge{}
	p.gauges[name] = g
	return g
}

func (p *mockProvider) Stop() {
}

func (c *mockCounter) Name() string {
	panic("not implemented")
}

type mockCounter struct {
	mtx     sync.Mutex
	lastAdd uint64
}

func (c *mockCounter) With(metrics.Field) metrics.Counter {
	panic("not implemented")
}

func (c *mockCounter) Add(delta uint64) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.lastAdd = delta

}

type mockGauge struct {
	mtx       sync.Mutex
	lastSet   float64
	lastDelta float64
}

func (g *mockGauge) Name() string {
	panic("not implemented")
}

func (g *mockGauge) With(metrics.Field) metrics.Gauge {
	panic("not implemented")
}

func (g *mockGauge) Set(value float64) {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.lastSet = value
}

func (g *mockGauge) Add(delta float64) {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.lastDelta = delta

}

func (g *mockGauge) Get() float64 {
	panic("not implemented")
}

type mockHistogram struct {
	mtx          sync.Mutex
	lastObserved int64
}

func (h *mockHistogram) Name() string {
	panic("not implemented")
}

func (h *mockHistogram) With(metrics.Field) metrics.Histogram {
	panic("not implemented")
}

func (h *mockHistogram) Observe(value int64) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.lastObserved = value
}

func (h *mockHistogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	panic("not implemented")
}
