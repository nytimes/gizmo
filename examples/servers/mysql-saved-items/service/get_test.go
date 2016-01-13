package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/server"
	"github.com/rcrowley/go-metrics"
)

// testSavedItemsRepo is a mock implementation of the SavedItemsRepo interface.
type testSavedItemsRepo struct {
	MockGet    func(uint64) ([]*SavedItem, error)
	MockPut    func(uint64, string) error
	MockDelete func(uint64, string) error
}

// Get will call the MockGet function of the test repo.
func (r *testSavedItemsRepo) Get(userID uint64) ([]*SavedItem, error) {
	return r.MockGet(userID)
}

// Put will call the MockPut function of the test repo.
func (r *testSavedItemsRepo) Put(userID uint64, url string) error {
	return r.MockPut(userID, url)
}

// Delete will call the MockDelete function of the test repo.
func (r *testSavedItemsRepo) Delete(userID uint64, url string) error {
	return r.MockDelete(userID, url)
}

func TestGet(t *testing.T) {
	testDate := time.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		givenID   string
		givenRepo func(uint64) ([]*SavedItem, error)

		wantCode  int
		wantError *jsonErr
		wantItems []*SavedItem
	}{
		{
			"123456",
			func(id uint64) ([]*SavedItem, error) {
				if id != 123456 {
					t.Errorf("mockget expected id of 123456; got %d", id)
				}
				return []*SavedItem{
					&SavedItem{
						123456,
						"http://nytimes.com/saved/item",
						testDate,
					},
				}, nil
			},

			http.StatusOK,
			nil,
			[]*SavedItem{
				&SavedItem{
					123456,
					"http://nytimes.com/saved/item",
					testDate,
				},
			},
		},
		{
			"",
			func(id uint64) ([]*SavedItem, error) {
				if id != 123456 {
					t.Errorf("mockget expected id of 123456; got %d", id)
				}
				return []*SavedItem{
					&SavedItem{
						123456,
						"http://nytimes.com/saved/item",
						testDate,
					},
				}, nil
			},

			http.StatusUnauthorized,
			UnauthErr,
			[]*SavedItem(nil),
		},
		{
			"123456",
			func(id uint64) ([]*SavedItem, error) {
				if id != 123456 {
					t.Errorf("mockget expected id of 123456; got %d", id)
				}
				return []*SavedItem{}, errors.New("nope")
			},

			http.StatusServiceUnavailable,
			ServiceUnavailableErr,
			[]*SavedItem(nil),
		},
	}

	for testnum, test := range tests {

		ss := server.NewSimpleServer(nil)
		testRepo := &testSavedItemsRepo{MockGet: test.givenRepo}
		sis := &SavedItemsService{testRepo}
		ss.Register(sis)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/svc/saved-items/user", nil)
		if test.givenID != "" {
			r.Header.Set("USER_ID", test.givenID)
		}

		ss.ServeHTTP(w, r)

		if w.Code != test.wantCode {
			t.Errorf("TEST[%d] expected status code of %d; got %d", testnum, test.wantCode, w.Code)
		}

		bod := w.Body.Bytes()
		if test.wantCode >= 300 {
			var gotErr *jsonErr
			json.Unmarshal(bod, &gotErr)
			if !reflect.DeepEqual(gotErr, test.wantError) {
				t.Errorf("TEST[%d] expected status response of '%#v'; got '%#v'", testnum, test.wantError, gotErr)
			}
		}

		var got []*SavedItem
		json.Unmarshal(bod, &got)
		if !reflect.DeepEqual(got, test.wantItems) {
			t.Errorf("TEST[%d] expected items of \n%#v; got \n%#v", testnum, test.wantItems, got)
		}

		// ** THIS IS REQUIRED in order to run the test multiple times.
		metrics.DefaultRegistry.UnregisterAll()
	}
}
