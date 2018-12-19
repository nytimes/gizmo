// +build go1.7

package server

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
	vars, _ := rawVars.(map[string]string)
	return vars
}

// SetRouteVars will set the given value into into the request context
// with the shared 'vars' storage key.
func SetRouteVars(r *http.Request, val interface{}) {
	if val == nil {
		return
	}

	r2 := r.WithContext(context.WithValue(r.Context(), varsKey, val))
	*r = *r2
}

type contextKey int

// key to set/retrieve URL params from a
// Gorilla request context.
const varsKey contextKey = 2
