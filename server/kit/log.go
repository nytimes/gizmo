package kit

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc/metadata"
)

// Logger will return a kit/log.Logger that has been injected
// into the context by the kit server. This function will only work
// within the scope of a request initiated by the server.
func Logger(ctx context.Context) log.Logger {
	return ctx.Value(logKey).(log.Logger)
}

// Log will pull the server log.Logger from the context and
// log the given keyvals with it.
func Log(ctx context.Context, keyvals ...interface{}) error {
	return Logger(ctx).Log(keyvals...)
}

// LoggerWithFields will pull any known request info from the context
// and include it into the log as key values.
func LoggerWithFields(ctx context.Context) log.Logger {
	l := Logger(ctx)
	// for HTTP requests
	l = withLog(ctx, l, http.ContextKeyRequestMethod, "http-method")
	l = withLog(ctx, l, http.ContextKeyRequestPath, "http-path")
	l = withLog(ctx, l, http.ContextKeyRequestURI, "http-uri")
	l = withLog(ctx, l, http.ContextKeyRequestXRequestID, "http-x-request-id")
	l = withLog(ctx, l, http.ContextKeyRequestRemoteAddr, "http-remote-addr")
	l = withLog(ctx, l, http.ContextKeyRequestXForwardedFor, "http-x-forwarded-for")
	l = withLog(ctx, l, http.ContextKeyRequestUserAgent, "http-user-agent")
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

// LogMsgWithFields will start with LoggerWithFields and then
// log the given message under the key "msg".
func LogMsgWithFields(ctx context.Context, msg string) error {
	return LoggerWithFields(ctx).Log("msg", msg)
}

// LogMsg will log the given message to the server logger
// with the key "msg".
func LogMsg(ctx context.Context, msg string) error {
	return Logger(ctx).Log("msg", msg)
}

// LogErrorMsgWithFields will start with LoggerWithFields and then log
// the given error under the key "error" and the given message under the key "msg".
func LogErrorMsgWithFields(ctx context.Context, err error, msg string) error {
	return Logger(ctx).Log("error", err, "msg", msg)
}

// LogErrorMsg will log the given error under the key "error" and the given message under
// the key "msg".
func LogErrorMsg(ctx context.Context, err error, msg string) error {
	return Logger(ctx).Log("error", err, "msg", msg)
}

func withLog(ctx context.Context, l log.Logger, key interface{}, skey string) log.Logger {
	if val := ctx.Value(key).(string); val != "" {
		return log.With(l, "http-method", val)
	}
	return l
}
