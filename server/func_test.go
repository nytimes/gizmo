package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

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
			SetRouteVars(r, mux.Vars(r))
			got := GetUInt64Var(r, "key")
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
			SetRouteVars(r, mux.Vars(r))
			got := GetInt64Var(r, "key")
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

		got, gotErr := ParseTruthyFalsy(test.given)

		if test.wantErr != (gotErr != nil) {
			t.Errorf("wantErr is %v, but got %s", test.wantErr, gotErr)
		}

		if test.want != got {
			t.Errorf("expected %v, got %v", test.want, got)

		}

	}

}
