package kit // import "github.com/NYTimes/gizmo/server/kit"

import (
	"os"
	"os/signal"
	"syscall"
)

// Run will use environment variables to configure the server then register the given
// Service and start up the server(s).
// This will block until the server shuts down.
func Run(service Service) error {
	svr := NewServer(service)

	if err := svr.start(); err != nil {
		return err
	}

	// parse address for host, port
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	svr.logger.Log("received signal", <-ch)
	return svr.stop()
}
