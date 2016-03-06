package appengine

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NYTimes/gizmo/appengineserver"
	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/examples/nyt/nyttest"
	"google.golang.org/appengine/aetest"
)

func TestAppEngineGetMostPopular(t *testing.T) {
	tests := []struct {
		givenURI    string
		givenClient *nyttest.CtxClient

		wantCode int
		wantBody interface{}
	}{
		{
			"/svc/nyt/most-popular/my-resource/my-section/1",
			&nyttest.CtxClient{
				TestGetMostPopular: func(resourceType, section string, timeframe uint) ([]*nyt.MostPopularResult, error) {
					if resourceType != "my-resource" {
						t.Errorf("expected resourceType of 'my-resource'; got %#v", resourceType)
					}
					if section != "my-section" {
						t.Errorf("expected section of 'my-section'; got %#v", section)
					}
					if timeframe != uint(1) {
						t.Errorf("expected timeframe of '1'; got %#v", timeframe)
					}
					return []*nyt.MostPopularResult{
						&nyt.MostPopularResult{
							Url: "https://www.nytimes.com/awesome-article",
							Id:  uint64(12345),
						},
					}, nil
				},
			},

			http.StatusOK,
			[]interface{}{
				map[string]interface{}{
					"id":  float64(12345),
					"url": "https://www.nytimes.com/awesome-article",
				},
			},
		},
		{
			"/svc/nyt/most-popular/my-resource/my-section/10",
			&nyttest.CtxClient{
				TestGetMostPopular: func(resourceType, section string, timeframe uint) ([]*nyt.MostPopularResult, error) {
					if resourceType != "my-resource" {
						t.Errorf("expected resourceType of 'my-resource'; got %#v", resourceType)
					}
					if section != "my-section" {
						t.Errorf("expected section of 'my-section'; got %#v", section)
					}
					if timeframe != uint(10) {
						t.Errorf("expected timeframe of '10'; got %#v", timeframe)
					}
					return nil, errors.New("unacceptable!!!")
				},
			},

			http.StatusServiceUnavailable,
			map[string]interface{}{
				"error": "sorry, this service is unavailable",
			},
		},
	}

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for _, test := range tests {

		srvr := appengineserver.NewSimpleServer(nil)
		nytclient = func() nyt.ContextClient { return test.givenClient }
		srvr.Register(&AppEngineService{})

		r, rerr := inst.NewRequest("GET", test.givenURI, nil)
		if rerr != nil {
			t.Fatalf("unable to create request: %v", rerr)
		}
		w := httptest.NewRecorder()
		srvr.ServeHTTP(w, r)

		if w.Code != test.wantCode {
			t.Errorf("expected response code of %d; got %d", test.wantCode, w.Code)
		}

		var got interface{}
		err := json.NewDecoder(w.Body).Decode(&got)
		if err != nil {
			t.Error("unable to JSON decode response body: ", err)
		}

		if !reflect.DeepEqual(got, test.wantBody) {
			t.Errorf("expected response body of\n%#v;\ngot\n%#v", test.wantBody, got)
		}

	}

}
