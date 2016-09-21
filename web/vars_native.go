// +build go1.7

package web

import (
	"context"
	"net/http"
)

// Vars is a helper function for accessing route
// parameters from any server.Router implementation. This is the equivalent
// of using `mux.Vars(r)` with the Gorilla mux.Router.
func Vars(r *http.Request) map[string]string {
	// vars doesnt exist yet, return empty map
	rawVars := r.Context().Value(varsKey)
	if rawVars == nil {
		return map[string]string{}
	}

	// for some reason, vars is wrong type, return empty map
	vars, ok := rawVars.(map[string]string)
	if !ok {
		return map[string]string{}
	}

	return vars
}

// SetRouteVars will set the given value into into the request context
// with the shared 'vars' storage key. This returns a shallow copy of
// the request with its context altered. Users will need to use the returned
// value for route vars to exist.
func SetRouteVars(r *http.Request, val interface{}) *http.Request {
	if val == nil {
		return r
	}

	return r.WithContext(context.WithValue(r.Context(), varsKey, val))
}

type contextKey int

// key to set/retrieve URL params from a
// Gorilla request context.
const varsKey contextKey = 2
