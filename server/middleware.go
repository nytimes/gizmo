package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// JSONToHTTP is the middleware func to convert a JSONEndpoint to
// an http.HandlerFunc.
func JSONToHTTP(ep JSONEndpoint) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer func() {
				if err := r.Body.Close(); err != nil {
					Log.Warn("unable to close request body: ", err)
				}
			}()
		}
		// it's JSON, so always set that content type
		w.Header().Set("Content-Type", jsonContentType)
		// prepare to grab the response from the ep
		var b bytes.Buffer
		encoder := json.NewEncoder(&b)

		// call the func and return err or not
		code, res, err := ep(r)
		w.WriteHeader(code)
		if err != nil {
			res = err
		}

		err = encoder.Encode(res)
		if err != nil {
			LogWithFields(r).Error("unable to JSON encode response: ", err)
		}

		if _, err := w.Write(b.Bytes()); err != nil {
			LogWithFields(r).Warn("unable to write response: ", err)
		}
	})
}

// CORSHandler is a middleware func for setting all headers that enable CORS.
// If an originSuffix is provided, a strings.HasSuffix check will be performed
// before adding any CORS header. If an empty string is provided, any Origin
// header found will be placed into the CORS header. If no Origin header is
// found, no headers will be added.
func CORSHandler(f http.Handler, originSuffix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" &&
			(originSuffix == "" || strings.HasSuffix(origin, originSuffix)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-requested-by, *")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		f.ServeHTTP(w, r)
	})
}

// NoCacheHandler is a middleware func for setting the Cache-Control to no-cache.
func NoCacheHandler(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		f.ServeHTTP(w, r)
	})
}

// JSONPHandler is a middleware func for wrapping response body with JSONP.
func JSONPHandler(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// using a custom ResponseWriter so we can
		// capture the response of the main request,
		// wrap our JSONP stuff around it
		// and only write to the actual response once
		jw := &jsonpResponseWriter{w: w}
		f.ServeHTTP(jw, r)

		// add the JSONP only if the callback exists
		callbackLabel := r.FormValue("callback")
		if callbackLabel != "" {
			var result []byte
			result = append(jsonpStart, []byte(callbackLabel)...)
			result = append(result, jsonpSecond...)
			result = append(result, jw.buf.Bytes()...)
			result = append(result, jsonpEnd...)
			if _, err := w.Write(result); err != nil {
				LogWithFields(r).Warn("unable to write JSONP response: ", err)
			}
		} else {
			// if no callback, just write the bytes
			if _, err := w.Write(jw.buf.Bytes()); err != nil {
				LogWithFields(r).Warn("unable to write response: ", err)
			}
		}
	})
}

var (
	jsonpStart  = []byte("/**/")
	jsonpSecond = []byte("(")
	jsonpEnd    = []byte(");")
)

type jsonpResponseWriter struct {
	w   http.ResponseWriter
	buf bytes.Buffer
}

func (w *jsonpResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *jsonpResponseWriter) WriteHeader(h int) {
	w.w.WriteHeader(h)
}

func (w *jsonpResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

// JSONContextToHTTP is a middleware func to convert a ContextHandler an http.Handler.
func JSONContextToHTTP(ep JSONContextEndpoint) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer func() {
				if err := r.Body.Close(); err != nil {
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
			LogWithFields(r).Error("unable to JSON encode response: ", err)
		}

		if _, err := w.Write(b.Bytes()); err != nil {
			LogWithFields(r).Warn("unable to write response: ", err)
		}
	})
}

// ContextToHTTP is a middleware func to convert a ContextHandler an http.Handler.
func ContextToHTTP(ep ContextHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ep.ServeHTTPContext(r.Context(), w, r)
	})
}

// WithCloseHandler returns a Handler cancelling the context when the client
// connection close unexpectedly.
func WithCloseHandler(h ContextHandler) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Cancel the context if the client closes the connection
		if wcn, ok := w.(http.CloseNotifier); ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			defer cancel()

			notify := wcn.CloseNotify()
			go func() {
				select {
				case <-notify:
					cancel()
				case <-ctx.Done():
				}
			}()
		}

		h.ServeHTTPContext(ctx, w, r)
	})
}
