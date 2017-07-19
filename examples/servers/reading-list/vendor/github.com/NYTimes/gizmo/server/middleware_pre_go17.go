// +build !go1.7

package server

import (
	"bytes"
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
)

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
		ep.ServeHTTPContext(context.Background(), w, r)
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
