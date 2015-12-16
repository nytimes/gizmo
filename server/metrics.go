package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rcrowley/go-metrics"
)

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
// digit of their HTTP status code via go-metrics.
type CounterByStatusXX struct {
	counter1xx, counter2xx, counter3xx, counter4xx, counter5xx metrics.Counter
	handler                                                    http.Handler
}

// CountedByStatusXX returns an http.Handler that passes requests to an
// underlying http.Handler and then counts the response by the first digit of
// its HTTP status code via go-metrics.
func CountedByStatusXX(handler http.Handler, name string, registry metrics.Registry) *CounterByStatusXX {
	if nil == registry {
		registry = metrics.DefaultRegistry
	}
	c := &CounterByStatusXX{
		counter1xx: metrics.NewCounter(),
		counter2xx: metrics.NewCounter(),
		counter3xx: metrics.NewCounter(),
		counter4xx: metrics.NewCounter(),
		counter5xx: metrics.NewCounter(),
		handler:    handler,
	}
	if err := registry.Register(
		fmt.Sprintf("%s-1xx", name),
		c.counter1xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-2xx", name),
		c.counter2xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-3xx", name),
		c.counter3xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-4xx", name),
		c.counter4xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-5xx", name),
		c.counter5xx,
	); nil != err {
		panic(err)
	}
	return c
}

// ServeHTTP passes the request to the underlying http.Handler and then counts
// the response by its HTTP status code via go-metrics.
func (c *CounterByStatusXX) ServeHTTP(w0 http.ResponseWriter, r *http.Request) {
	w := newMetricsResponseWriter(w0)
	c.handler.ServeHTTP(w, r)
	if w.StatusCode < 200 {
		c.counter1xx.Inc(1)
	} else if w.StatusCode < 300 {
		c.counter2xx.Inc(1)
	} else if w.StatusCode < 400 {
		c.counter3xx.Inc(1)
	} else if w.StatusCode < 500 {
		c.counter4xx.Inc(1)
	} else {
		c.counter5xx.Inc(1)
	}
}

// Timer is an http.Handler that counts requests via go-metrics.
type Timer struct {
	metrics.Timer
	handler http.Handler
}

// Timed returns an http.Handler that starts a timer, passes requests to an
// underlying http.Handler, stops the timer, and updates the timer via
// go-metrics.
func Timed(handler http.Handler, name string, registry metrics.Registry) *Timer {
	timer := &Timer{
		Timer:   metrics.NewTimer(),
		handler: handler,
	}
	if nil == registry {
		registry = metrics.DefaultRegistry
	}
	if err := registry.Register(name, timer); nil != err {
		panic(err)
	}
	return timer
}

// ServeHTTP starts a timer, passes the request to the underlying http.Handler,
// stops the timer, and updates the timer via go-metrics.
func (t *Timer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer t.UpdateSince(time.Now())
	t.handler.ServeHTTP(w, r)
}

/*
The contents of this file are derived from the 'https://github.com/rcrowley/go-tigertonic' package.

Copyright 2013 Richard Crowley. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

    1.  Redistributions of source code must retain the above copyright
        notice, this list of conditions and the following disclaimer.

    2.  Redistributions in binary form must reproduce the above
        copyright notice, this list of conditions and the following
        disclaimer in the documentation and/or other materials provided
        with the distribution.

THIS SOFTWARE IS PROVIDED BY RICHARD CROWLEY ``AS IS'' AND ANY EXPRESS
OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL RICHARD CROWLEY OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF
THE POSSIBILITY OF SUCH DAMAGE.

The views and conclusions contained in the software and documentation
are those of the authors and should not be interpreted as representing
official policies, either expressed or implied, of Richard Crowley.
*/
