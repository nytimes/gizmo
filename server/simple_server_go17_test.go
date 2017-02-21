// +build go1.7

package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NYTimes/gizmo/web"
)

type benchmarkContextService struct {
	fast bool
}

func (s *benchmarkContextService) Prefix() string {
	return "/svc/v1"
}

func (s *benchmarkContextService) ContextEndpoints() map[string]map[string]ContextHandlerFunc {
	return map[string]map[string]ContextHandlerFunc{
		"/ctx/1/{something}/:something": map[string]ContextHandlerFunc{
			"GET": s.GetSimple,
		},
		"/ctx/2": map[string]ContextHandlerFunc{
			"GET": s.GetSimpleNoParam,
		},
	}
}

func (s *benchmarkContextService) ContextMiddleware(h ContextHandler) ContextHandler {
	return h
}

func (s *benchmarkContextService) Middleware(h http.Handler) http.Handler {
	return h
}

func (s *benchmarkContextService) GetSimple(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	something := web.Vars(r)["something"]
	fmt.Fprint(w, something)
}

func (s *benchmarkContextService) GetSimpleNoParam(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
