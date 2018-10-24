package kit

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc/metadata"
)

// Logger will return a kit/log.Logger that has been injected into the context by the kit
// server. This logger has had request headers and metadata added as key values.
// This function will only work within the scope of a request initiated by the server.
func Logger(ctx context.Context) log.Logger {
	return ctx.Value(logKey).(log.Logger)
}

// Log will pull a request scoped log.Logger from the context and
// log the given keyvals with it.
func Log(ctx context.Context, keyvals ...interface{}) error {
	return Logger(ctx).Log(keyvals...)
}

// AddLogKeyVals will add any common HTTP headers or gRPC metadata
// from the given context to the given logger as fields.
// This is used by the server to initialize the request scopes logger.
func AddLogKeyVals(ctx context.Context, l log.Logger) log.Logger {
	// for HTTP requests
	keys := map[interface{}]string{
		http.ContextKeyRequestMethod:        "http-method",
		http.ContextKeyRequestURI:           "http-uri",
		http.ContextKeyRequestPath:          "http-path",
		http.ContextKeyRequestHost:          "http-host",
		http.ContextKeyRequestXRequestID:    "http-x-request-id",
		http.ContextKeyRequestRemoteAddr:    "http-remote-addr",
		http.ContextKeyRequestXForwardedFor: "http-x-forwarded-for",
		http.ContextKeyRequestUserAgent:     "http-user-agent",
		ContextKeyCloudTraceContext:         cloudTraceLogKey,
	}
	for k, v := range keys {
		if val, ok := ctx.Value(k).(string); ok && val != "" {
			l = log.With(l, v, val)
		}
	}
	// for gRPC requests
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return l
	}
	for k, v := range md {
		l = log.With(l, k, v)
	}
	return l
}

// LogMsg will log the given message to the server logger
// with the key "message" along with all the common request headers or gRPC metadata.
func LogMsg(ctx context.Context, message string) error {
	return Logger(ctx).Log("message", message)
}

// LogErrorMsg will log the given error under the key "error", the given message under
// the key "message" along with all the common request headers or gRPC metadata.
func LogErrorMsg(ctx context.Context, err error, message string) error {
	return Logger(ctx).Log("error", err, "message", message)
}
