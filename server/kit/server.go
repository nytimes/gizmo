// +build !appengine

package kit

// TODO(jprobinson): built in stackdriver tracing (sampling)

// Run will use environment variables to configure the server then register the given
// Service and start up the server(s). The ready channel will be closed once the
// service has started.
// Run will block until the quit channel is closed.
func Run(service Service, ready chan struct{}, quit chan struct{}) error {
	svr := NewServer(service)
	if err := svr.start(); err != nil {
		return err
	}

	close(ready)

	_ = <-quit
	svr.logger.Log("received quit message")
	return svr.stop()
}
