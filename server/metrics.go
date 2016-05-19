package server

import (
	"bufio"
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

// metricsResponseWriter grabs the StatusCode.
type metricsResponseWriter struct {
	w          http.ResponseWriter
	StatusCode int
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{w: w}
}

func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *metricsResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *metricsResponseWriter) WriteHeader(h int) {
	w.StatusCode = h
	w.w.WriteHeader(h)
}

func (w *metricsResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not implement hijacker")
	}
	return h.Hijack()
}

// CounterByStatusXX is an http.Handler that counts responses by the first
// digit of their HTTP status code via go-kit/kit/metrics.
type CounterByStatusXX struct {
	counter1xx, counter2xx, counter3xx, counter4xx, counter5xx metrics.Counter
	handler                                                    http.Handler
}

// CountedByStatusXX returns an http.Handler that passes requests to an
// underlying http.Handler and then counts the response by the first digit of
// its HTTP status code via go-kit/kit/metrics.
func CountedByStatusXX(handler http.Handler, name string, p provider.Provider) *CounterByStatusXX {
	return &CounterByStatusXX{
		counter1xx: p.NewCounter(fmt.Sprintf("%s-1xx", name),
			fmt.Sprintf("counting responses to %q with a status code of 1XX", name)),
		counter2xx: p.NewCounter(fmt.Sprintf("%s-2xx", name),
			fmt.Sprintf("counting responses to %q with a status code of 2XX", name)),
		counter3xx: p.NewCounter(fmt.Sprintf("%s-3xx", name),
			fmt.Sprintf("counting responses to %q with a status code of 3XX", name)),
		counter4xx: p.NewCounter(fmt.Sprintf("%s-4xx", name),
			fmt.Sprintf("counting responses to %q with a status code of 4XX", name)),
		counter5xx: p.NewCounter(fmt.Sprintf("%s-5xx", name),
			fmt.Sprintf("counting responses to %q with a status code of 5XX", name)),
		handler: handler,
	}
}

// ServeHTTP passes the request to the underlying http.Handler and then counts
// the response by its HTTP status code via go-kit/kit/metrics.
func (c *CounterByStatusXX) ServeHTTP(w0 http.ResponseWriter, r *http.Request) {
	w := newMetricsResponseWriter(w0)
	c.handler.ServeHTTP(w, r)
	if w.StatusCode < 200 {
		c.counter1xx.Add(1)
	} else if w.StatusCode < 300 {
		c.counter2xx.Add(1)
	} else if w.StatusCode < 400 {
		c.counter3xx.Add(1)
	} else if w.StatusCode < 500 {
		c.counter4xx.Add(1)
	} else {
		c.counter5xx.Add(1)
	}
}

// Timer is an http.Handler that counts requests via go-kit/kit/metrics.
type Timer struct {
	metrics.TimeHistogram
	handler http.Handler
}

// Timed returns an http.Handler that starts a timer, passes requests to an
// underlying http.Handler, stops the timer, and updates the timer via
// go-kit/kit/metrics.
func Timed(handler http.Handler, name string, p provider.Provider) *Timer {
	hist, err := p.NewHistogram(name,
		fmt.Sprintf("tracking request duration for %q", name),
		0, 1500000, 4, // 0-15 minute time range, 4 sigfigs
		50, 75, 90, 95, 99) // quantiles
	if err != nil {
		panic("invalid histogram settings")
	}

	return &Timer{
		TimeHistogram: metrics.NewTimeHistogram(time.Millisecond, hist),
		handler:       handler,
	}
}

// ServeHTTP starts a timer, passes the request to the underlying http.Handler,
// stops the timer, and updates the timer via go-kit/kit/metrics.
func (t *Timer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) { t.Observe(time.Since(start)) }(time.Now())
	t.handler.ServeHTTP(w, r)
}
