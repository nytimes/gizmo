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

func TestGetCats(t *testing.T) {
	tests := []struct {
		givenURI    string
		givenClient *nyttest.CtxClient

		wantCode int
		wantBody interface{}
	}{
		{
			"/svc/nyt/cats",
			&nyttest.CtxClient{
				TestSemanticConceptSearch: func(conceptType, concept string) ([]*nyt.SemanticConceptArticle, error) {
					return []*nyt.SemanticConceptArticle{
						&nyt.SemanticConceptArticle{
							Url: "https://www.nytimes.com/awesome-article",
						},
					}, nil
				},
			},

			http.StatusOK,
			[]interface{}{
				map[string]interface{}{
					"url": "https://www.nytimes.com/awesome-article",
				},
			},
		},
		{
			"/svc/nyt/cats",
			&nyttest.CtxClient{
				TestSemanticConceptSearch: func(conceptType, concept string) ([]*nyt.SemanticConceptArticle, error) {
					return nil, errors.New("NOPE!")
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

		r, _ := inst.NewRequest("GET", test.givenURI, nil)
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
