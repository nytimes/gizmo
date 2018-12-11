package kit

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc/metadata"
)

// NewLogger will inspect the environment and, if running in the Google App Engine
// environment, it will return a new Stackdriver logger annotated with the current
// server's project ID, service ID and version. If not in App Engine, a normal JSON
// logger pointing to stdout will be returned.
// This function can be used for services that need to log information outside the
// context of an inbound request.
// When using the Stackdriver logger, any go-kit/log/levels will be translated to
// Stackdriver severity levels.
func NewLogger(ctx context.Context) (log.Logger, func() error, error) {
	// running locally or in a non-GAE environment? use JSON
	if !isGAE() {
		return log.NewJSONLogger(log.NewSyncWriter(os.Stdout)), func() error { return nil }, nil
	}

	projectID, serviceID, svcVersion := getGAEInfo()
	return newAppEngineLogger(ctx, projectID, serviceID, svcVersion)
}

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
