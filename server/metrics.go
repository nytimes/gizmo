package server

import (
	"bufio"
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
	kprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		counter1xx: p.NewCounter(fmt.Sprintf("%s-1xx", name)),
		counter2xx: p.NewCounter(fmt.Sprintf("%s-2xx", name)),
		counter3xx: p.NewCounter(fmt.Sprintf("%s-3xx", name)),
		counter4xx: p.NewCounter(fmt.Sprintf("%s-4xx", name)),
		counter5xx: p.NewCounter(fmt.Sprintf("%s-5xx", name)),
		handler:    handler,
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
	metrics.Histogram
	handler http.Handler
}

// Timed returns an http.Handler that starts a timer, passes requests to an
// underlying http.Handler, stops the timer, and updates the timer via
// go-kit/kit/metrics.
func Timed(handler http.Handler, name string, p provider.Provider) *Timer {
	return &Timer{
		Histogram: p.NewHistogram(name, 50),
		handler:   handler,
	}
}

// ServeHTTP starts a timer, passes the request to the underlying http.Handler,
// stops the timer, and updates the timer via go-kit/kit/metrics.
func (t *Timer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) {
		t.Observe(time.Since(start).Seconds())
	}(time.Now())
	t.handler.ServeHTTP(w, r)
}

func metricName(fullPath, method string) string {
	// replace slashes
	fullPath = strings.Replace(fullPath, "/", "-", -1)
	// replace periods
	fullPath = strings.Replace(fullPath, ".", "-", -1)
	return fmt.Sprintf("routes.%s-%s", fullPath, method)
}

// TimedAndCounted wraps a http.Handler with Timed and a CountedByStatusXXX or, if the
// metrics provider is of type Prometheus, via a prometheus.InstrumentHandler
func TimedAndCounted(handler http.Handler, fullPath string, method string, p provider.Provider) *Timer {
	fullPath = strings.TrimPrefix(fullPath, "/")
	switch fmt.Sprintf("%T", p) {
	case "*provider.prometheusProvider":
		return PrometheusTimedAndCounted(handler, fullPath)
	default:
		mn := metricName(fullPath, method)
		return Timed(CountedByStatusXX(handler, mn+".STATUS-COUNT", p), mn+".DURATION", p)
	}
}

// PrometheusTimedAndCounted wraps a http.Handler with via prometheus.InstrumentHandler
func PrometheusTimedAndCounted(handler http.Handler, name string) *Timer {
	// TODO: add better regex to find and replace
	// only [a-zA-Z_] is allowed in prometheus metric names
	bads := []string{"{", "}", "[", "]", ":", ".", "-", "?", "!", "+"}
	mname := strings.Replace(name, "/", "_", -1)
	for _, bad := range bads {
		mname = strings.Replace(mname, bad, "", -1)
	}
	// since we cant use summaries or histograms in stackdriver,
	// setup gauges for the quantiles we care about
	pcts := []float64{0.05, 0.50, 0.90, 0.95, 0.99}
	gs := map[float64]metrics.Gauge{}
	for _, p := range pcts {
		gname := fmt.Sprintf("%s_%2.2f_request_duration", mname, p)
		// remove the decimal if there
		gname = strings.Replace(gname, ".", "_", -1)
		gs[p] = kprom.NewGaugeFrom(
			prometheus.GaugeOpts{
				Name: gname,
				Help: fmt.Sprintf("Request latency for %s's %2.2f percentile", name, p),
			},
			nil,
		)
	}
	// and kick off go goroutine to update histo with gauges in background
	// sure, why not?
	histo := generic.NewHistogram(name, 50)
	go func() {
		c := time.Tick(20 * time.Second)
		for _ = range c {
			for p, g := range gs {
				v := histo.Quantile(p)
				if v > 0 {
					g.Set(v)
				}
			}
		}
	}()
	// create a request code counter
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: mname + "_requests_total",
			Help: "Total Requests for " + name,
		},
		[]string{"code", "method"},
	)
	prometheus.MustRegister(counter)
	return &Timer{
		Histogram: histo,
		handler: promhttp.InstrumentHandlerCounter(counter,
			handler,
		),
	}
}
