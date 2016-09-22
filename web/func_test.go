package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/web"
	"github.com/gorilla/mux"
)

func TestParseISODate(t *testing.T) {
	tests := []struct {
		given string

		want    time.Time
		wantErr bool
	}{
		{
			"2015-10-29",

			time.Date(2015, time.October, 29, 0, 0, 0, 0, time.Local),
			false,
		},
		{
			"20151029",

			time.Time{},
			true,
		},
	}

	for _, test := range tests {
		got, gotErr := web.ParseISODate(test.given)

		if test.wantErr {
			if gotErr == nil {
				t.Error("expected an error and did not get one")
			}
			continue
		}

		if gotErr != nil {
			t.Error("did not expect an error but got one: ", gotErr)
		}

		if !got.Equal(test.want) {
			t.Errorf("got %#v, expected %#v: ", got, test.want)
		}

	}
}

func TestParseDateRangeFullDay(t *testing.T) {
	tests := []struct {
		given map[string]string

		wantStart time.Time
		wantEnd   time.Time
		wantErr   bool
	}{
		{
			map[string]string{
				"startDate": "2015-10-29",
				"endDate":   "2015-10-31",
			},

			time.Date(2015, time.October, 29, 0, 0, 0, 0, time.Local),
			time.Date(2015, time.October, 31, 23, 59, 59, 1000, time.Local),
			false,
		},
		{
			map[string]string{
				"startDate": "2015-10-29",
				"endDate":   "-10-31",
			},

			time.Time{},
			time.Time{},
			true,
		},
		{
			map[string]string{
				"startDate": "-10-29",
				"endDate":   "2015-10-31",
			},

			time.Time{},
			time.Time{},
			true,
		},
		{
			map[string]string{
				"endDate": "2015-10-31",
			},

			time.Time{},
			time.Time{},
			true,
		},
		{
			map[string]string{
				"startDate": "-10-29",
			},

			time.Time{},
			time.Time{},
			true,
		},
	}

	for _, test := range tests {
		gotStart, gotEnd, gotErr := web.ParseDateRangeFullDay(test.given)

		if test.wantErr {
			if gotErr == nil {
				t.Error("expected an error and did not get one")
			}
			continue
		}

		if gotErr != nil {
			t.Error("did not expect an error but got one: ", gotErr)
		}

		if !gotStart.Equal(test.wantStart) {
			t.Errorf("got start date of %#v, expected %#v: ", gotStart, test.wantStart)
		}

		if !gotEnd.Equal(test.wantEnd) {
			t.Errorf("got end date of %#v, expected %#v: ", gotStart, test.wantStart)
		}
	}
}

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		given map[string]string

		wantStart time.Time
		wantEnd   time.Time
		wantErr   bool
	}{
		{
			map[string]string{
				"startDate": "2015-10-29",
				"endDate":   "2015-10-31",
			},

			time.Date(2015, time.October, 29, 0, 0, 0, 0, time.Local),
			time.Date(2015, time.October, 31, 0, 0, 0, 0, time.Local),
			false,
		},
		{
			map[string]string{
				"startDate": "2015-10-29",
				"endDate":   "-10-31",
			},

			time.Time{},
			time.Time{},
			true,
		},
		{
			map[string]string{
				"startDate": "-10-29",
				"endDate":   "2015-10-31",
			},

			time.Time{},
			time.Time{},
			true,
		},
		{
			map[string]string{
				"endDate": "2015-10-31",
			},

			time.Time{},
			time.Time{},
			true,
		},
		{
			map[string]string{
				"startDate": "-10-29",
			},

			time.Time{},
			time.Time{},
			true,
		},
	}

	for _, test := range tests {
		gotStart, gotEnd, gotErr := web.ParseDateRange(test.given)

		if test.wantErr {
			if gotErr == nil {
				t.Error("expected an error and did not get one")
			}
			continue
		}

		if gotErr != nil {
			t.Error("did not expect an error but got one: ", gotErr)
		}

		if !gotStart.Equal(test.wantStart) {
			t.Errorf("got start date of %#v, expected %#v: ", gotStart, test.wantStart)
		}

		if !gotEnd.Equal(test.wantEnd) {
			t.Errorf("got end date of %#v, expected %#v: ", gotStart, test.wantStart)
		}
	}
}

func TestGetUInt64Var(t *testing.T) {

	tests := []struct {
		givenURL   string
		givenRoute string

		want uint64
	}{
		{
			"/blah/123",
			"/blah/{key}",
			123,
		},
		{
			"/blah/adsf",
			"/blah/{key}",
			0,
		},
		{
			"/blah?key=123",
			"/blah",
			123,
		},
		{
			"/blah?key=abc",
			"/blah",
			0,
		},
	}

	for _, test := range tests {
		route := mux.NewRouter()
		route.HandleFunc(test.givenRoute, func(w http.ResponseWriter, r *http.Request) {
			web.SetRouteVars(r, mux.Vars(r))
			got := web.GetUInt64Var(r, "key")
			if got != test.want {
				t.Errorf("URL(%s): got int of %#v, expected %#v", test.givenURL, got, test.want)
			}
		})

		r, _ := http.NewRequest("GET", test.givenURL, nil)
		route.ServeHTTP(httptest.NewRecorder(), r)
	}
}

func TestGetInt64Var(t *testing.T) {

	tests := []struct {
		givenURL   string
		givenRoute string

		want int64
	}{
		{
			"/blah/123",
			"/blah/{key}",
			123,
		},
		{
			"/blah/adsf",
			"/blah/{key}",
			0,
		},
		{
			"/blah?key=123",
			"/blah",
			123,
		},
		{
			"/blah?key=abc",
			"/blah",
			0,
		},
	}

	for _, test := range tests {
		route := mux.NewRouter()
		route.HandleFunc(test.givenRoute, func(w http.ResponseWriter, r *http.Request) {
			web.SetRouteVars(r, mux.Vars(r))
			got := web.GetInt64Var(r, "key")
			if got != test.want {
				t.Errorf("URL(%s): got int of %#v, expected %#v", test.givenURL, got, test.want)
			}
		})

		r, _ := http.NewRequest("GET", test.givenURL, nil)
		route.ServeHTTP(httptest.NewRecorder(), r)
	}
}

func TestParseTruthyFalsy(t *testing.T) {
	tests := []struct {
		given interface{}

		want    bool
		wantErr bool
	}{
		{
			"true",
			true,
			false,
		},
		{
			"false",
			false,
			false,
		},
		{
			"0",
			false,
			false,
		},
		{
			"1",
			true,
			false,
		},
		{
			"",
			false,
			false,
		},
		{
			"nope!",
			false,
			true,
		},
		{
			1,
			true,
			false,
		},
		{
			0,
			false,
			false,
		},
		{
			2,
			false,
			true,
		},
		{
			float64(0.0),
			false,
			false,
		},
		{
			float64(1.0),
			true,
			false,
		},
		{
			float64(2.0),
			false,
			true,
		},
		{
			true,
			true,
			false,
		},
		{
			false,
			false,
			false,
		},
	}

	for _, test := range tests {

		got, gotErr := web.ParseTruthyFalsy(test.given)

		if test.wantErr != (gotErr != nil) {
			t.Errorf("wantErr is %v, but got %s", test.wantErr, gotErr)
		}

		if test.want != got {
			t.Errorf("expected %v, got %v", test.want, got)

		}

	}

}
