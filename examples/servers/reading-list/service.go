package readinglist

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/trace"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gizmo/server/kit"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type service struct {
	db     DB
	tracer *trace.Client
}

// ensure we implement the gRPC service
var _ ReadingListServiceServer = &service{}

func NewService(db DB) (kit.Service, error) {
	ctx := context.Background()
	pid := os.Getenv("GCP_PROJECT_ID")

	tracer, err := trace.NewClient(ctx, pid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init trace client")
	}

	return &service{
		db:     db,
		tracer: tracer,
	}, nil
}

func (s *service) HTTPOptions() []httptransport.ServerOption {
	return []httptransport.ServerOption{}
}

// override the default gorilla router and select the stdlib
func (s *service) HTTPRouterOptions() []kit.RouterOption {
	return []kit.RouterOption{
		kit.RouterSelect("stdlib"),
	}
}

// in this example, we're tossing a simple CORS middleware in the mix
func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	return server.CORSHandler(h, "")
}

// the go-kit middleware is used for checking user authentication and
// injecting the current user into the request context.
func (s *service) Middleware(ep endpoint.Endpoint) endpoint.Endpoint {
	return endpoint.Endpoint(func(ctx context.Context, r interface{}) (interface{}, error) {
		start := time.Now()
		defer func() {
			kit.LogMsgWithFields(ctx,
				fmt.Sprintf("complete in %0.8f seconds", time.Since(start).Seconds()))
		}()

		usr, err := getUserFromMD(ctx)
		if usr == "" || err != nil {
			kit.LogErrorMsgWithFields(ctx, err, "unable to get user auth")
			// reject if user is not logged in
			return nil, kit.NewJSONStatusResponse(
				&Message{"please provide a valid oauth token"},
				http.StatusUnauthorized,
			)
		}
		// add the user to the request context and continue
		return ep(addUser(ctx, usr), r)
	})
}

// declare the endpoints for the HTTP server
func (s *service) HTTPEndpoints() map[string]map[string]kit.HTTPEndpoint {
	return map[string]map[string]kit.HTTPEndpoint{
		"/link": {
			"PUT": {
				Endpoint: s.putLink,
				Decoder:  decodePutRequest,
			},
		},
		"/list/{limit:[0-9]+}": {
			"GET": {
				Endpoint: s.getLinks,
				Decoder:  decodeGetRequest,
			},
		},
	}
}

func (s *service) RPCMiddleware() grpc.UnaryServerInterceptor {
	if s.tracer != nil {
		return s.tracer.GRPCServerInterceptor()
	}
	return nil
}

func (s *service) RPCServiceDesc() *grpc.ServiceDesc {
	return &_ReadingListService_serviceDesc
}

func (s *service) RPCOptions() []grpc.ServerOption {
	return nil
}

const userKey = "oauth-user"

func getUserFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no request metadata")
	}

	infos, ok := md["x-endpoint-api-userinfo"]
	if !ok || len(infos) == 0 {
		return "", errors.New("no user info")
	}

	js, err := base64.StdEncoding.DecodeString(infos[0])
	if err != nil {
		return "", errors.Wrap(err, "invalid user info")
	}

	usr := struct {
		ID string `json:"id"`
	}{}
	err = json.Unmarshal(js, &usr)
	return usr.ID, errors.Wrap(err, "unable to decode user info")
}

func addUser(ctx context.Context, usr string) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

func getUser(ctx context.Context) string {
	return ctx.Value(userKey).(string)
}
