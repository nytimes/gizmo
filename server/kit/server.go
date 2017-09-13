// +build !appengine

package kit

import (
	"os"
	"os/signal"
	"syscall"
)

// TODO(jprobinson): built in stackdriver tracing (sampling)

// Run will use environment variables to configure the server then register the given
// Service and start up the server(s). The ready channel will be closed once the
// service has started.
// Run will block until the quit channel is closed.
func Run(service Service, ready chan struct{}, quit chan struct{}, errors chan error) {
	defer close(errors)
	svr := NewServer(service)
	if err := svr.start(); err != nil {
		errors <- err
		close(ready)
		return
	}

	signals := make(chan os.Signal, 1)
	defer close(signals)

	close(ready)
	svr.logger.Log("server is ready - closed ready channel")

	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	// wait for quit channel or os signal
	select {
	case <-signals:
		svr.logger.Log("received os quit signal")
	case <-quit:
		svr.logger.Log("received quit message")
	}

	err := svr.stop()
	if err != nil {
		errors <- err
	} else {
		svr.logger.Log("stopped server")
	}
}
