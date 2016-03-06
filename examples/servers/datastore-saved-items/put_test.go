package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NYTimes/gizmo/appengineserver"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/user"
)

func TestAppEnginePut(t *testing.T) {

	tests := []struct {
		givenID   string
		givenURL  string
		givenRepo func(string, string) error

		wantCode  int
		wantError *jsonErr
		wantResp  *jsonResponse
	}{
		{
			"123456",
			"http://nytimes.com/article",
			func(id string, url string) error {
				if id != "123456" {
					t.Errorf("MockPut expected id of 123456; got %s", id)
				}
				if url != "http://nytimes.com/article" {
					t.Errorf("MockPut expected url of `http://nytimes.com/aritcle'; got %s", url)
				}
				return nil
			},

			http.StatusCreated,
			&jsonErr{},
			&jsonResponse{"successfully saved item"},
		},
		{
			"123456",
			"http://nytimes.com/article",
			func(id string, url string) error {
				if id != "123456" {
					t.Errorf("MockPut expected id of 123456; got %s", id)
				}
				if url != "http://nytimes.com/article" {
					t.Errorf("MockPut expected url of `http://nytimes.com/aritcle'; got %s", url)
				}
				return errors.New("nope")
			},

			http.StatusServiceUnavailable,
			ServiceUnavailableErr,
			&jsonResponse{""},
		},
		{
			"",
			"http://nytimes.com/article",
			func(id string, url string) error {
				t.Error("MockPut should not have been called in this scenario!")
				return nil
			},

			http.StatusFound,
			nil,
			nil,
		},
	}

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for _, test := range tests {

		// create a new Gizmo simple server
		ss := appengineserver.NewSimpleServer(nil)
		// create our test repo implementation
		testRepo := &testSavedItemsRepo{MockPut: test.givenRepo}
		// inject the test repo into a new SavedItemsService
		sis := &SavedItemsService{testRepo}
		// register the service with our simple server
		ss.Register(sis)

		// set up the w and r to pass into our server
		w := httptest.NewRecorder()
		r, _ := inst.NewRequest("PUT", "/svc/saved-items?url="+test.givenURL, nil)
		if test.givenID != "" {
			aetest.Login(&user.User{Email: "eml", ID: test.givenID}, r)
		}

		// run the test by passing a request we expect to hit our endpoint
		// into the simple server's ServeHTTP method.
		ss.ServeHTTP(w, r)

		// first test validation: check the HTTP response code
		if w.Code != test.wantCode {
			t.Errorf("expected status code of %d; got %d", test.wantCode, w.Code)
		}

		// get the body of the response to inspect
		bod := w.Body.Bytes()

		// if we were expecting an error scenario, marshal the response
		// JSON into an error to compare with what we want.
		var gotErr *jsonErr
		json.Unmarshal(bod, &gotErr)
		if !reflect.DeepEqual(gotErr, test.wantError) {
			t.Errorf("expected status response of '%#v'; got '%#v'", test.wantError, gotErr)
		}

		// if we expect a normal response, compare it to our wanted response struct
		var got *jsonResponse
		json.Unmarshal(bod, &got)
		if !reflect.DeepEqual(got, test.wantResp) {
			t.Errorf("expected response of \n%#v; got \n%#v", test.wantResp, got)
		}
	}
}
