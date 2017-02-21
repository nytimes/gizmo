// +build go1.7

package server

import (
	"context"
	"net/http"
)

// JSONContextEndpoint is the JSONContextService equivalent to JSONService's JSONEndpoint.
type JSONContextEndpoint func(context.Context, *http.Request) (int, interface{}, error)

// ContextHandlerFunc is an equivalent to SimpleService's http.HandlerFunc.
type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPContext is an implementation of ContextHandler interface.
func (h ContextHandlerFunc) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	h(ctx, rw, req)
}

// ContextHandler is an equivalent to http.Handler but with additional param.
type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}
