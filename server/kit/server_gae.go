// +build appengine

package kit

import "net/http"

// Run will not actually start a server if in the App Engine environment.
func Run(service Service) error {
	http.Handle("/", NewServer(service))
	return nil
}
