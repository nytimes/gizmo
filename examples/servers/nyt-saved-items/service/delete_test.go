package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NYTimes/gizmo/server"
	"github.com/rcrowley/go-metrics"
)

func TestDelete(t *testing.T) {

	tests := []struct {
		givenHeaders http.Header
		givenURL     string
		givenRepo    func(uint64, string) error

		wantCode  int
		wantError *jsonErr
		wantResp  *jsonResponse
	}{
		{
			http.Header{"USER_ID": []string{"123456"}},
			"http://nytimes.com/article",
			func(id uint64, url string) error {
				if id != 123456 {
					t.Errorf("MockDelete expected id of 123456; got %d", id)
				}
				if url != "http://nytimes.com/article" {
					t.Errorf("MockDelete expected url of `http://nytimes.com/aritcle'; got %s", url)
				}
				return nil
			},

			http.StatusOK,
			nil,
			&jsonResponse{"successfully deleted saved item"},
		},
		{
			http.Header{"USER_ID": []string{"123456"}},
			"http://nytimes.com/article",
			func(id uint64, url string) error {
				if id != 123456 {
					t.Errorf("MockDelete expected id of 123456; got %d", id)
				}
				if url != "http://nytimes.com/article" {
					t.Errorf("MockDelete expected url of `http://nytimes.com/aritcle'; got %s", url)
				}
				return errors.New("nope!")
			},

			http.StatusServiceUnavailable,
			ServiceUnavailableErr,
			&jsonResponse{},
		},
		{
			http.Header{},
			"http://nytimes.com/article",
			func(id uint64, url string) error {
				t.Error("MockDelete should not have been called in this scenario!")
				return nil
			},

			http.StatusUnauthorized,
			UnauthErr,
			&jsonResponse{},
		},
	}

	for _, test := range tests {

		ss := server.NewSimpleServer(nil)
		testRepo := &testSavedItemsRepo{MockDelete: test.givenRepo}
		sis := &SavedItemsService{testRepo}
		ss.Register(sis)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("DELETE", "/svc/saved-items/user?url="+test.givenURL, nil)
		r.Header = test.givenHeaders

		ss.ServeHTTP(w, r)

		if w.Code != test.wantCode {
			t.Errorf("expected status code of %d; got %d", test.wantCode, w.Code)
		}

		bod := w.Body.Bytes()
		if test.wantCode >= 300 {
			var gotErr *jsonErr
			json.Unmarshal(bod, &gotErr)
			if !reflect.DeepEqual(gotErr, test.wantError) {
				t.Errorf("expected status response of '%#v'; got '%#v'", test.wantError, gotErr)
			}
		}

		var got *jsonResponse
		json.Unmarshal(bod, &got)
		if !reflect.DeepEqual(got, test.wantResp) {
			t.Errorf("expected response of \n%#v; got \n%#v", test.wantResp, got)
		}

		// ** THIS IS REQUIRED in order to run the test multiple times.
		metrics.DefaultRegistry.UnregisterAll()
	}
}
