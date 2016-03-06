package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/appengineserver"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/user"
)

// testSavedItemsRepo is a mock implementation of the SavedItemsRepo interface.
type testSavedItemsRepo struct {
	MockGet    func(string) ([]*SavedItem, error)
	MockPut    func(string, string) error
	MockDelete func(string, string) error
}

// Get will call the MockGet function of the test repo.
func (r *testSavedItemsRepo) Get(ctx context.Context, userID string) ([]*SavedItem, error) {
	return r.MockGet(userID)
}

// Put will call the MockPut function of the test repo.
func (r *testSavedItemsRepo) Put(ctx context.Context, userID, url string) error {
	return r.MockPut(userID, url)
}

// Delete will call the MockDelete function of the test repo.
func (r *testSavedItemsRepo) Delete(ctx context.Context, userID, url string) error {
	return r.MockDelete(userID, url)
}

func TestGet(t *testing.T) {
	testDate := time.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		givenID   string
		givenRepo func(string) ([]*SavedItem, error)

		wantCode  int
		wantError *jsonErr
		wantItems []*SavedItem
	}{
		{
			"123456",
			func(id string) ([]*SavedItem, error) {
				if id != "123456" {
					t.Errorf("mockget expected id of 123456; got %s", id)
				}
				return []*SavedItem{
					&SavedItem{
						"123456",
						"http://nytimes.com/saved/item",
						testDate,
					},
				}, nil
			},

			http.StatusOK,
			&jsonErr{},
			[]*SavedItem{
				&SavedItem{
					"123456",
					"http://nytimes.com/saved/item",
					testDate,
				},
			},
		},
		{
			"",
			func(id string) ([]*SavedItem, error) {
				if id != "123456" {
					t.Errorf("mockget expected id of 123456; got %s", id)
				}
				return []*SavedItem{
					&SavedItem{
						"123456",
						"http://nytimes.com/saved/item",
						testDate,
					},
				}, nil
			},

			http.StatusFound,
			nil,
			nil,
		},
		{
			"123456",
			func(id string) ([]*SavedItem, error) {
				if id != "123456" {
					t.Errorf("mockget expected id of 123456; got %s", id)
				}
				return []*SavedItem{}, errors.New("nope")
			},

			http.StatusServiceUnavailable,
			ServiceUnavailableErr,
			[]*SavedItem(nil),
		},
	}

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for testnum, test := range tests {

		// create a new Gizmo simple server
		ss := appengineserver.NewSimpleServer(nil)
		// create our test repo implementation
		testRepo := &testSavedItemsRepo{MockGet: test.givenRepo}
		// inject the test repo into a new SavedItemsService
		sis := &SavedItemsService{testRepo}
		// register the service with our simple server
		ss.Register(sis)

		// set up the w and r to pass into our server
		w := httptest.NewRecorder()
		r, _ := inst.NewRequest("GET", "/svc/saved-items", nil)
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
		var got []*SavedItem
		json.Unmarshal(bod, &got)
		if !reflect.DeepEqual(got, test.wantItems) {
			t.Errorf("TEST[%d] expected items of \n%#v; got \n%#v", testnum, test.wantItems, got)
		}
	}
}
