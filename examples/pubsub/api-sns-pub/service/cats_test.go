package service

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	"github.com/NYTimes/gizmo/server"
	"github.com/golang/protobuf/proto"
)

func TestGetCats(t *testing.T) {
	tests := []struct {
		given      string
		givenError error

		wantCode      int
		wantPublished []nyt.SemanticConceptArticle
	}{
		{
			`{"url":"http://www.nytiems.com/cats-article","title":"cats cats cats"}`,
			nil,

			http.StatusOK,
			[]nyt.SemanticConceptArticle{
				nyt.SemanticConceptArticle{
					Url:   "http://www.nytiems.com/cats-article",
					Title: "cats cats cats",
				},
			},
		},
		{
			`{"url":"http://www.nytiems.com/cats-article","title":"cats cats cats"}`,
			errors.New("NOPE!"),

			http.StatusServiceUnavailable,
			[]nyt.SemanticConceptArticle{
				nyt.SemanticConceptArticle{
					Url:   "http://www.nytiems.com/cats-article",
					Title: "cats cats cats",
				},
			},
		},
		{
			`"cats cats cats"`,
			nil,

			http.StatusServiceUnavailable,
			[]nyt.SemanticConceptArticle{},
		},
	}

	for _, test := range tests {

		srvr := server.NewSimpleServer(nil)
		pub := &pubsubtest.TestPublisher{GivenError: test.givenError}
		srvr.Register(&JSONPubService{pub: pub})

		r, _ := http.NewRequest("PUT", "/svc/nyt/cats", bytes.NewBufferString(test.given))
		w := httptest.NewRecorder()
		srvr.ServeHTTP(w, r)

		if w.Code != test.wantCode {
			t.Errorf("expected response code of %d; got %d", test.wantCode, w.Code)
		}

		if len(pub.Published) != len(test.wantPublished) {
			t.Errorf("expected %d published messages, got %d: ", len(pub.Published), len(test.wantPublished))
		}

		for idx, wantPub := range test.wantPublished {
			var got nyt.SemanticConceptArticle
			proto.Unmarshal(pub.Published[idx].Body, &got)
			if !reflect.DeepEqual(got, wantPub) {
				t.Errorf("expected response body of\n%#v;\ngot\n%#v", wantPub, got)
			}
		}
	}

}
