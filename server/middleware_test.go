package server

import (
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
