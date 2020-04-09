package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCORSHandler(t *testing.T) {
	tests := []struct {
		given       string
		givenPrefix string

		wantOrigin  string
		wantCreds   string
		wantHeaders string
		wantMethods string
	}{
		{
			"",
			"",
			"",
			"",
			"",
			"",
		},
		{
			".nytimes.com.",
			"",
			".nytimes.com.",
			"true",
			"Content-Type, x-requested-by, *",
			"GET, PUT, POST, DELETE, OPTIONS",
		},
		{
			".nytimes.com.",
			"blah.com",
			"",
			"",
			"",
			"",
		},
	}

	for _, test := range tests {
		r, _ := http.NewRequest("GET", "", nil)
		r.Header.Add("Origin", test.given)
		w := httptest.NewRecorder()

		CORSHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), test.givenPrefix).ServeHTTP(w, r)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != test.wantOrigin {
			t.Errorf("expected CORS origin header to be '%#v', got '%#v'", test.wantOrigin, got)
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != test.wantCreds {
			t.Errorf("expected CORS creds header to be '%#v', got '%#v'", test.wantCreds, got)
		}
		if got := w.Header().Get("Access-Control-Allow-Headers"); got != test.wantHeaders {
			t.Errorf("expected CORS 'headers' header to be '%#v', got '%#v'", test.wantHeaders, got)
		}
		if got := w.Header().Get("Access-Control-Allow-Methods"); got != test.wantMethods {
			t.Errorf("expected CORS 'methods' header to be '%#v', got '%#v'", test.wantMethods, got)
		}
	}
}

func TestJSONToHTTP(t *testing.T) {
	tests := []struct {
		given     JSONEndpoint
		givenBody io.Reader

		wantCode int
		wantBody string
	}{
		{
			JSONEndpoint(func(r *http.Request) (int, interface{}, error) {
				bod, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Error("unable to read given request body: ", err)
				}
				if string(bod) != "yup" {
					t.Errorf("expected 'yup', got %+v", string(bod))
				}
				return http.StatusOK, struct{ Howdy string }{"Hi"}, nil
			}),
			bytes.NewBufferString("yup"),
			http.StatusOK,
			"{\"Howdy\":\"Hi\"}\n",
		},
		{
			JSONEndpoint(func(r *http.Request) (int, interface{}, error) {
				return http.StatusServiceUnavailable, nil, &testJSONError{"nope"}
			}),
			nil,
			http.StatusServiceUnavailable,
			"{\"error\":\"nope\"}\n",
		},
	}

	for _, test := range tests {
		r, _ := http.NewRequest("GET", "", test.givenBody)
		w := httptest.NewRecorder()
		JSONToHTTP(test.given).ServeHTTP(w, r)

		if w.Code != test.wantCode {
			t.Errorf("expected status code %d, got %d", test.wantCode, w.Code)
		}
		if gotHdr := w.Header().Get("Content-Type"); gotHdr != jsonContentType {
			t.Errorf("expected Content-Type header of '%#v', got '%#v'", jsonContentType, gotHdr)
		}
		if got := w.Body.String(); got != test.wantBody {
			t.Errorf("expected body of '%#v', got '%#v'", test.wantBody, got)
		}
	}
}

type testJSONError struct {
	Err string `json:"error"`
}

func (t *testJSONError) Error() string {
	return t.Err
}

func TestJSONPHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	r.Form = url.Values{"callback": {"harumph"}}
	w := httptest.NewRecorder()

	JSONPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"jsonp\":\"sucks\"}"))
	})).ServeHTTP(w, r)

	want := `/**/harumph({"jsonp":"sucks"});`
	if got := w.Body.String(); got != want {
		t.Errorf("expected JSONP response of '%#v', got '%#v'", want, got)
	}

	// once again, without a callback
	r, _ = http.NewRequest("GET", "", nil)
	w = httptest.NewRecorder()

	JSONPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"jsonp\":\"sucks\"}"))
	})).ServeHTTP(w, r)

	want = `{"jsonp":"sucks"}`
	if got := w.Body.String(); got != want {
		t.Errorf("expected JSONP response of '%#v', got '%#v'", want, got)
	}
}

func TestNoCacheHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	NoCacheHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	want := "no-cache, no-store, must-revalidate"
	if got := w.Header().Get("Cache-Control"); got != want {
		t.Errorf("expected no-cache control header to be '%#v', got '%#v'", want, got)
	}
	want = "no-cache"
	if got := w.Header().Get("Pragma"); got != want {
		t.Errorf("expected no-cache pragma header to be '%#v', got '%#v'", want, got)
	}
	want = "0"
	if got := w.Header().Get("Expires"); got != want {
		t.Errorf("expected no-cache Expires header to be '%#v', got '%#v'", want, got)
	}
}

func TestAppIDHandler(t *testing.T) {
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Error("failed to create mock request", "err", err)
	}
	w := httptest.NewRecorder()

	id := "flambe"
	AppIDHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		// write the ID to the response body so we can test it on the recorder
		_, _ = w.Write([]byte(requestID))
		w.WriteHeader(http.StatusOK)
	}), &MockIDer{sendThis: id}).ServeHTTP(w, r)

	headVal, ok := w.Result().Header[RequestIDHeader]
	if !ok {
		t.Error("header value was not found")
	}
	if len(headVal) != 1 {
		t.Error("expected one value in request ID header", "got", len(headVal))
	}
	if headVal[0] != id {
		t.Error("unexpected value in request ID header", "got", headVal[0], "expected", id)
	}
	if w.Body.String() != id {
		t.Error("unexpected value in body", "got", w.Body.String(), "expected", id)
	}
}

func TestPipelineIDHandler(t *testing.T) {
	tests := []struct {
		desc, prev, next, expected string
	}{
		{"EmptyPrev", "", "roger", "roger"},
		{"FirstConcat", "roger", "roderick", "roger|roderick"},
		{"SecondConcat", "roger|roderick", "brian", "roger|roderick|brian"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {

			r, err := http.NewRequest("GET", "", nil)
			if err != nil {
				t.Error("failed to create mock request", "err", err)
			}
			r.Header.Set(RequestIDHeader, test.prev)
			w := httptest.NewRecorder()

			pipeIDer := &PipelineID{
				AppIDer: &MockIDer{sendThis: test.next},
			}

			PipelineIDHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID := GetRequestID(r.Context())
				// write the ID to the response body so we can test it on the recorder
				_, _ = w.Write([]byte(requestID))
				w.WriteHeader(http.StatusOK)
			}), pipeIDer).ServeHTTP(w, r)

			headVal, ok := w.Result().Header[RequestIDHeader]
			if !ok {
				t.Error("header value was not found")
			}
			if len(headVal) != 1 {
				t.Error("expected one value in request ID header", "got", len(headVal))
			}
			if headVal[0] != test.expected {
				t.Error("unexpected value in request ID header", "got", headVal[0], "expected", test.expected)
			}
			if w.Body.String() != test.expected {
				t.Error("unexpected value in body", "got", w.Body.String(), "expected", test.expected)
			}
		})
	}
}
