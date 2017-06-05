package kit

import (
	"context"

	"github.com/go-kit/kit/log"
)

// Logger will return a kit/log.Logger that has been injected
// into the context by the kit server. This function will only work
// within the scope of a request initiated by the server.
func Logger(ctx context.Context) log.Logger {
	return ctx.Value(logKey).(log.Logger)
}
