package kit

import (
	"context"

	"github.com/go-kit/kit/log"
)

func Logger(ctx context.Context) log.Logger {
	return ctx.Value(logKey).(log.Logger)
}
