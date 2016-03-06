package appengineserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// JSONToHTTPContext is the middleware func to convert a JSONEndpoint to
// an http.HandlerFunc.
func JSONToHTTPContext(ep JSONEndpoint) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer func() {
				if err := r.Body.Close(); err != nil {
					log.Warningf(ctx, "unable to close request body: ", err)
				}
			}()
		}
		// it's JSON, so always set that content type
		w.Header().Set("Content-Type", jsonContentType)
		// prepare to grab the response from the ep
		var b bytes.Buffer
		encoder := json.NewEncoder(&b)

		// call the func and return err or not
		code, res, err := ep(ctx, r)
		w.WriteHeader(code)
		if err != nil {
			res = err
		}

		err = encoder.Encode(res)
		if err != nil {
			log.Errorf(ctx, "unable to JSON encode response: ", err)
		}

		if _, err := w.Write(b.Bytes()); err != nil {
			log.Warningf(ctx, "unable to write response: ", err)
		}
	})
}

// ContextToHTTP is a middleware func to convert a ContextHandler an http.Handler.
func ContextToHTTP(ep ContextHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ep.ServeHTTPContext(appengine.NewContext(r), w, r)
	})
}

// CORSHandler is a middleware func for setting all headers that enable CORS.
// If an originSuffix is provided, a strings.HasSuffix check will be performed
// before adding any CORS header. If an empty string is provided, any Origin
// header found will be placed into the CORS header. If no Origin header is
// found, no headers will be added.
func CORSHandler(h ContextHandler, originSuffix string) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" &&
			(originSuffix == "" || strings.HasSuffix(origin, originSuffix)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, *")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
		}
		h.ServeHTTPContext(ctx, w, r)
	})
}

// NoCacheHandler is a middleware func for setting the Cache-Control to no-cache.
func NoCacheHandler(h ContextHandler) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		h.ServeHTTPContext(ctx, w, r)
	})
}
