package server

import "net/http"

func httpServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:        handler,
		MaxHeaderBytes: maxHeaderBytes,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
	}
}
