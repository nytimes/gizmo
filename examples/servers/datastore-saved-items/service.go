package service

import (
	"fmt"
	"net/http"

	"github.com/NYTimes/gizmo/appengineserver"
	"github.com/NYTimes/gziphandler"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func init() {
	appengineserver.Init(nil, NewSavedItemsService())
}

// SavedItemsService will keep a handle on the saved items repository and implement
// the gizmo/server.JSONService interface.
type SavedItemsService struct {
	repo SavedItemsRepo
}

// NewSavedItemsService will attempt to instantiate a new repository and service.
func NewSavedItemsService() *SavedItemsService {
	return &SavedItemsService{NewSavedItemsRepo()}
}

// Prefix is to implement gizmo/server.Service interface. The string will be prefixed to all endpoint
// routes.
func (s *SavedItemsService) Prefix() string {
	return "/svc"
}

// Middleware provides a hook to add service-wide http.Handler middleware to the service.
// In this example we are using it to add GZIP compression to our responses.
// This method helps satisfy the server.Service interface.
func (s *SavedItemsService) Middleware(h http.Handler) http.Handler {
	// wrap the response with our GZIP Middleware
	return gziphandler.GzipHandler(h)
}

func (s *SavedItemsService) ContextMiddleware(h appengineserver.ContextHandler) appengineserver.ContextHandler {
	return appengineserver.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if usr := user.Current(ctx); usr == nil {
			s.LoginRedirect(ctx, w, r)
			return
		}
		h.ServeHTTPContext(ctx, w, r)
	})
}

// JSONMiddleware provides a hook to add service-wide middleware for how JSONEndpoints
// should behave. In this example, we’re using the hook to check for a header to
// identify and authorize the user. This method helps satisfy the server.JSONService interface.
func (s *SavedItemsService) JSONMiddleware(j appengineserver.JSONEndpoint) appengineserver.JSONEndpoint {
	return func(ctx context.Context, r *http.Request) (code int, res interface{}, err error) {

		// call the endpoint
		code, res, err = j(ctx, r)

		// if the endpoint returns an unexpected error, return a generic message
		// and log it.
		if err != nil && code != http.StatusUnauthorized {
			// LogWithFields will add all the request context values
			// to the structured log entry along some other request info
			log.Warningf(ctx, "unexpected service error: %s", err)
			return http.StatusServiceUnavailable, nil, ServiceUnavailableErr
		}

		return code, res, err
	}
}

// JSONEndpoints is the most important method of the Service implementation. It provides a
// listing of all endpoints available in the service with their routes and HTTP methods.
// This method helps satisfy the server.JSONService interface.
func (s *SavedItemsService) JSONEndpoints() map[string]map[string]appengineserver.JSONEndpoint {
	return map[string]map[string]appengineserver.JSONEndpoint{
		"/saved-items": map[string]appengineserver.JSONEndpoint{
			"GET":    s.Get,
			"PUT":    s.Put,
			"POST":   s.Put,
			"DELETE": s.Delete,
		},
	}
}

func (s *SavedItemsService) Endpoints() map[string]map[string]appengineserver.ContextHandlerFunc {
	return map[string]map[string]appengineserver.ContextHandlerFunc{
		"/login": map[string]appengineserver.ContextHandlerFunc{
			"GET": s.LoginRedirect,
		},
		"/add": map[string]appengineserver.ContextHandlerFunc{
			"GET": s.ViewSave,
		},
	}
}

func (s *SavedItemsService) ViewSave(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<html>
<head></head>
<body>
	<form action="/svc/saved-items" method="POST">	
		<input type="text" name="url"></input>
		<input type="submit" value="save url" />
	</form>
</body>
</html>`)
}

func (s *SavedItemsService) LoginRedirect(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	url, err := user.LoginURL(ctx, "https://nyt-reading-list.appspot.com/svc/saved-items")
	if err != nil {
		log.Criticalf(ctx, "unable to obtain login url: ", err)
		http.Error(w, "nope", http.StatusServiceUnavailable)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

type (
	// jsonResponse is a generic struct for responding with a simple JSON message.
	jsonResponse struct {
		Message string `json:"message"`
	}
	// jsonErr is a tiny helper struct to make displaying errors in JSON better.
	jsonErr struct {
		Err string `json:"error"`
	}
)

func (e *jsonErr) Error() string {
	return e.Err
}

var (
	// ServiceUnavailableErr is a global error that will get returned when we are experiencing
	// technical issues.
	ServiceUnavailableErr = &jsonErr{"sorry, this service is currently unavailable"}
	// UnauthErr is a global error returned when the user does not supply the proper
	// authorization headers.
	UnauthErr = &jsonErr{"please include a valid USER_ID header in the request"}
)
