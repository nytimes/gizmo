package kit

import (
	"context"
	"os"

	"github.com/NYTimes/gizmo/observe"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc/metadata"
)

// NewLogger will inspect the environment and, if running in the Google App Engine,
// Google Kubernetes Engine, Google Compute Engine or AWS EC2 environment,
// it will return a new Stackdriver logger annotated with the current
// server's project ID, service ID and version and other environment specific values.
// If not in App Engine, GKE, GCE or AWS EC2 - a normal JSON logger pointing to stdout
// will be returned.
// This function can be used for services that need to log information outside the
// context of an inbound request.
// When using the Stackdriver logger, any go-kit/log/levels will be translated to
// Stackdriver severity levels.
// The logID field is used when the server is deployed in a Stackdriver enabled environment.
// If an empty string is provided, "gae_log" will be used in App Engine and "stdout" elsewhere.
// For more information about to use of logID see the documentation here: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#FIELDS.log_name
func NewLogger(ctx context.Context, logID string) (log.Logger, func() error, error) {
	projectID, serviceID, svcVersion := observe.GetServiceInfo()
	lg, cl, err := newStackdriverLogger(ctx, logID, projectID, serviceID, svcVersion)
	// if Stackdriver logger was not able to find information about monitored resource it returns nil.
	if err != nil {
		// running locally or in a non-GAE environment? use JSON
		lg := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		lg.Log("error", err,
			"message", "unable to initialize Stackdriver logger. falling back to stdout JSON logging.")
		return lg, func() error { return nil }, nil
	}
	return lg, cl, err
}

// SetLogger sets log.Logger to the context and returns new context with logger.
func SetLogger(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, logKey, logger)
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
