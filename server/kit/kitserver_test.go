package kit_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	ocontext "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/NYTimes/gziphandler"
)

func TestKitServer(t *testing.T) {
	shutdownErrChan := make(chan error)
	go func() {
		// runs the HTTP _AND_ gRPC servers
		shutdownErrChan <- kit.Run(&server{})
	}()
	// server may still be coming up, give it 3 attempts
	var (
		err  error
		resp *http.Response
	)
	for i := 0; i < 3; i++ {
		// hit the health check
		resp, err = http.Get("http://localhost:8080/healthz")
		if err == nil {
			break
		}
		t.Logf("healthcheck failed on attempt %d", i+1)
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatal("unable to hit health check:", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("unable to read health check response:", err)
	}

	if string(b) != "OK" {
		t.Fatalf("unexpected health check response. got %q, wanted 'OK'", string(b))
	}

	// hit the HTTP server
	resp, err = http.Get("http://localhost:8080/svc/cat/ziggy")
	if err != nil {
		t.Fatal("unable to cat http endpoint:", err)
	}

	var hcat Cat
	err = json.NewDecoder(resp.Body).Decode(&hcat)
	if err != nil {
		t.Fatal("unable to read JSON cat:", err)
	}

	if !reflect.DeepEqual(&hcat, testCat) {
		t.Fatalf("expected cat: %#v, got %#v", testCat, hcat)
	}

	// hit the gRPC server
	gc, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("unable to init gRPC connection: %s", err)
	}
	defer gc.Close()
	cc := NewKitTestServiceClient(gc)
	cat, err := cc.GetCatName(context.Background(), &GetCatNameRequest{Name: "ziggy"})
	if err != nil {
		t.Fatalf("unable to make gRPC request: %s", err)
	}

	if !reflect.DeepEqual(cat, testCat) {
		t.Fatalf("expected cat: %#v, got %#v", testCat, cat)
	}

	// make signal to kill server
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	t.Log("waiting for shutdown")
	err = <-shutdownErrChan
	t.Log("shutdown complete")
	if err != nil {
		t.Fatal("problems running service: " + err.Error())
	}
}

type server struct{}

func (s *server) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return endpoint.Endpoint(func(ctx context.Context, r interface{}) (interface{}, error) {
		kit.LogMsg(ctx, "kit middleware!")
		return e(ctx, r)
	})
}

func (s *server) HTTPMiddleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

func (s *server) HTTPOptions() []httptransport.ServerOption {
	return nil
}

func (s *server) HTTPRouterOptions() []kit.RouterOption {
	return nil
}

func (s *server) HTTPEndpoints() map[string]map[string]kit.HTTPEndpoint {
	return map[string]map[string]kit.HTTPEndpoint{
		"/svc/cat/{name:[a-zA-Z]+}": {
			"GET": {
				Endpoint: s.getCatByName,
				Decoder:  catNameDecoder,
			},
		},
	}
}

func (s *server) RPCServiceDesc() *grpc.ServiceDesc {
	return &_KitTestService_serviceDesc
}

func (s *server) RPCMiddleware() grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx ocontext.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		kit.LogMsg(ctx, "rpc middleware!")
		return handler(ctx, req)
	})
}

func (s *server) RPCOptions() []grpc.ServerOption {
	return nil
}

// gRPC layer
func (s *server) GetCatName(ctx ocontext.Context, r *GetCatNameRequest) (*Cat, error) {
	rs, err := s.getCatByName(ctx, r)
	if err != nil {
		return nil, err
	}
	return rs.(*Cat), nil
}

// http decoder layer
func catNameDecoder(ctx context.Context, r *http.Request) (interface{}, error) {
	return &GetCatNameRequest{Name: kit.Vars(r)["name"]}, nil
}

var testCat = &Cat{Breed: "American Shorthair", Name: "Ziggy", Age: 12}

// shared business layer
func (s *server) getCatByName(ctx context.Context, _ interface{}) (interface{}, error) {
	kit.Logger(ctx).Log("message", "getting ziggy")
	return testCat, nil
}
