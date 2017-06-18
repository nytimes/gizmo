package kit_test

import (
	"io/ioutil"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/examples/servers/kit/api"
	"github.com/NYTimes/gizmo/server/kit"
	"github.com/kelseyhightower/envconfig"
)

func TestKitServer(t *testing.T) {
	var cfg api.Config
	envconfig.MustProcess("", &cfg)

	go func() {
		// runs the HTTP _AND_ gRPC servers
		err := kit.Run(api.New(cfg))
		if err != nil {
			t.Fatal("problems running service: " + err.Error())
		}
	}()

	// let the server start
	time.Sleep(1 * time.Second)

	// hit the health check
	resp, err := http.Get("http://localhost:8080/healthz")
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

	// make signal to kill server
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
