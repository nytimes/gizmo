package gcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/auth"
	"github.com/go-kit/kit/log"
	"golang.org/x/oauth2"
)

func TestAuthCallback(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2019, 9, 23, 21, 0, 0, 0, time.UTC) }
	auth.TimeNow = timeNow
	keyServer, authServer := setupAuthenticatorTest(t)
	defer keyServer.Close()
	defer authServer.Close()

	auth, err := NewAuthenticator(context.Background(), AuthenticatorConfig{
		CookieName: "example-cookie",
		IDConfig: IdentityConfig{
			Audience: "http://example.com",
			CertURL:  keyServer.URL,
		},
		IDVerifyFunc: func(_ context.Context, cs IdentityClaimSet) bool {
			if cs.Aud != "http://example.com" {
				return false
			}
			return strings.HasPrefix(cs.Email, "auth-example@")
		},
		AuthConfig: &oauth2.Config{
			RedirectURL: "http://localhost/oauthcallback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  authServer.URL,
				TokenURL: authServer.URL,
			},
		},
		Logger:      log.NewJSONLogger(os.Stdout),
		UnsafeState: true,
	})
	if err != nil {
		t.Fatalf("unable to init authenticator: %s", err)
	}

	var passedAuth bool
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedAuth = true
		w.WriteHeader(http.StatusOK)
	}))

	// make a call to out callback endpoint that has no state info
	r := httptest.NewRequest(http.MethodGet, "/randoendpoint", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if passedAuth {
		t.Fatal("request passed the auth layer despite having no known token")
	}

	got := w.Result()
	if got.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected to be get a 307 but got a status of %d instead",
			got.StatusCode)
	}

	// try to get the callback to play nice with an added state
	r = httptest.NewRequest(http.MethodGet,
		"/oauthcallback?state=eyJFeHBpcnkiOiIyMDE5LTA5LTIzVDE3OjUwOjQxLjYxOTc4Ny0wNDowMCIsIlVSSSI6Ii9yYW5kb2VuZHBvaW50IiwiTm9uY2UiOlsxNzUsOTIsMjUzLDQxLDg5LDIzMSwxNTAsMjQyLDk4LDY0LDY4LDE4NSwyMzMsMTM2LDcyLDIwOCwwLDIxLDIzLDg0LDEyMywxMzUsMTM5LDk2XX0=&code=XYZ",
		nil)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if passedAuth {
		t.Fatal("request passed the auth layer despite having no known token")
	}

	got = w.Result()

	if got.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected to be get a 200 OK but got a status of %d instead",
			got.StatusCode)
	}

	gotCookie := got.Header.Get("Set-Cookie")
	if gotCookie == "" {
		t.Fatal("expected cookie to have been dropped but got none")
	}
	cookieVals := strings.Split(strings.Split(gotCookie, "; ")[0], "=")
	if len(cookieVals) != 2 {
		t.Fatalf("cookie has unexpected format: %q", gotCookie)
	}
	if cookieVals[1] != testAuthToken {
		t.Fatalf("expected testAuthToken (%q), got %q", testAuthToken, cookieVals[1])
	}
}

func TestAuthenticatorTokenReject(t *testing.T) {
	timeNow = func() time.Time {
		return time.Date(2019, 9, 23, 22, 0, 0, 0, time.UTC)
	}
	auth.TimeNow = timeNow

	keyServer, authServer := setupAuthenticatorTest(t)
	defer keyServer.Close()
	defer authServer.Close()

	auth, err := NewAuthenticator(context.Background(), AuthenticatorConfig{
		CookieName: "example-cookie",
		IDConfig: IdentityConfig{
			Audience: "http://example.com",
			CertURL:  keyServer.URL,
		},
		IDVerifyFunc: func(_ context.Context, cs IdentityClaimSet) bool {
			return false // reject _all_ the things
		},
		AuthConfig: &oauth2.Config{
			Endpoint: oauth2.Endpoint{
				AuthURL:  authServer.URL,
				TokenURL: authServer.URL,
			},
		},
		Logger:      log.NewJSONLogger(os.Stdout),
		UnsafeState: true,
	})
	if err != nil {
		t.Fatalf("unable to init authenticator: %s", err)
	}

	var passedAuth bool
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedAuth = true
		w.WriteHeader(http.StatusOK)
	}))

	// add our known token to the outbound request
	r := httptest.NewRequest(http.MethodGet, "/bobloblaw", nil)
	r.Header.Set("Authorization", "Bearer "+testAuthToken)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if passedAuth {
		t.Fatal("request passed the auth layer but all requests should be rejected")
	}

	if got := w.Result(); got.StatusCode != http.StatusForbidden {
		t.Fatalf("expected to be get a 403 Forbidden but got a status of %d instead",
			got.StatusCode)
	}
}

func TestAuthenticatorTokenSuccess(t *testing.T) {
	timeNow = func() time.Time {
		return time.Date(2019, 9, 23, 22, 0, 0, 0, time.UTC)
	}
	auth.TimeNow = timeNow

	keyServer, authServer := setupAuthenticatorTest(t)
	defer keyServer.Close()
	defer authServer.Close()

	auth, err := NewAuthenticator(context.Background(), AuthenticatorConfig{
		CookieName: "example-cookie",
		IDConfig: IdentityConfig{
			Audience: "http://example.com",
			CertURL:  keyServer.URL,
		},
		IDVerifyFunc: func(_ context.Context, cs IdentityClaimSet) bool {
			if cs.Aud != "http://example.com" {
				return false
			}
			return strings.HasPrefix(cs.Email, "auth-example@")
		},
		AuthConfig: &oauth2.Config{
			Endpoint: oauth2.Endpoint{
				AuthURL:  authServer.URL,
				TokenURL: authServer.URL,
			},
		},
		Logger:      log.NewJSONLogger(os.Stdout),
		UnsafeState: true,
	})
	if err != nil {
		t.Fatalf("unable to init authenticator: %s", err)
	}

	var passedAuth bool
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedAuth = true
		w.WriteHeader(http.StatusOK)
	}))

	// add our known token to the outbound request
	r := httptest.NewRequest(http.MethodGet, "/bobloblaw", nil)
	r.Header.Set("Authorization", "Bearer "+testAuthToken)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !passedAuth {
		t.Fatal("request did not pass the auth layer despite having known token")
	}

	if got := w.Result(); got.StatusCode != http.StatusOK {
		t.Fatalf("expected to be get a 200 OK but got a status of %d instead",
			got.StatusCode)
	}
	// reset for next run
	passedAuth = false

	// add the same token as a cookie to also verify that plays nice
	r = httptest.NewRequest(http.MethodGet, "/bobloblaw", nil)
	r.Header.Set("Cookie", "example-cookie="+testAuthToken)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !passedAuth {
		t.Fatal("request did not pass the auth layer despite having known token within cookie")
	}

	if got := w.Result(); got.StatusCode != http.StatusOK {
		t.Fatalf("expected to be get a 200 OK but got a status of %d instead",
			got.StatusCode)
	}
}

const testAuthToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjBiMGJmMTg2NzQzNDcxYTFlZGNhYzMwNjBkMTI1NmY5ZTQwNTBiYTgiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhdWQiOiJodHRwOi8vZXhhbXBsZS5jb20iLCJhenAiOiJhdXRoLWV4YW1wbGVAbnl0LWdvbGFuZy1kZXYuaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20iLCJzdWIiOiIxMDMzNTk3OTYyODUxOTI5NzE4NzQiLCJlbWFpbCI6ImF1dGgtZXhhbXBsZUBueXQtZ29sYW5nLWRldi5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpYXQiOjE1NjkyNzIxNjgsImV4cCI6MTU2OTI3NTc2OH0.YCnNzU8mw_bdHmpmAWjcRc8NKs2A2ugz2XenN3opyEddKl9UxnMx-Y7k3Hd5jIhIZbBLp5_nwUojiWSoWXIYrIG-63MNINUCyoZykxwWMXhQTvTChPk69j0ex0wvwfuR044GrH1SRohYZET5JnlfrBroHjSOK0OqHjpePBp84ezK7EXwnKTgvqTB_lTp5__Xmwguw1DkLKVH9lpnU9RalAdjQZL0_tsK3MWSrVrL8byqP7MyOF6t5Xv-Xrb90feZIuJITPDtNoLvxL-ZXN5B-oGVyBlDK3w6mwTjLV4YQCa5lZKy3SrVHgAa4ucFkZFw0kzCJEnRY_YLkGh7c9eh2w"

func TestAuthCustomExceptions(t *testing.T) {
	timeNow = func() time.Time {
		return time.Date(2019, 9, 23, 0, 0, 0, 0, time.UTC)
	}
	auth.TimeNow = timeNow

	keyServer, authServer := setupAuthenticatorTest(t)
	defer keyServer.Close()
	defer authServer.Close()

	auth, err := NewAuthenticator(context.Background(), AuthenticatorConfig{
		IDConfig: IdentityConfig{
			Audience: "example.com",
			CertURL:  keyServer.URL,
		},
		CustomExceptionsFunc: func(_ context.Context, r *http.Request) bool {
			return r.URL.Path == "/bobloblaw"
		},
		AuthConfig: &oauth2.Config{
			Endpoint: oauth2.Endpoint{
				AuthURL:  authServer.URL,
				TokenURL: authServer.URL,
			},
		},
		UnsafeState: true,
	})
	if err != nil {
		t.Fatalf("unable to init authenticator: %s", err)
	}

	var passedAuth bool
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedAuth = true
		w.WriteHeader(http.StatusOK)
	}))

	// hit once without the special path, expect a redirect/no pass
	r := httptest.NewRequest(http.MethodGet, "/xyz", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if passedAuth {
		t.Fatal("request passed the auth layer without hitting the special path")
	}

	if got := w.Result(); got.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected to be redirected but got a status of %d instead",
			got.StatusCode)
	}

	// use special path, expect to get through
	r = httptest.NewRequest(http.MethodGet, "/bobloblaw", nil)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !passedAuth {
		t.Fatal("request did not pass the auth layer despite hitting the special path")
	}

	if got := w.Result(); got.StatusCode != http.StatusOK {
		t.Fatalf("expected to be get a 200 OK but got a status of %d instead",
			got.StatusCode)
	}

}

func TestAuthHeaderExceptions(t *testing.T) {
	timeNow = func() time.Time {
		return time.Date(2019, 9, 23, 0, 0, 0, 0, time.UTC)
	}
	auth.TimeNow = timeNow

	keyServer, authServer := setupAuthenticatorTest(t)
	defer keyServer.Close()
	defer authServer.Close()

	auth, err := NewAuthenticator(context.Background(), AuthenticatorConfig{
		IDConfig: IdentityConfig{
			Audience: "example.com",
			CertURL:  keyServer.URL,
		},
		HeaderExceptions: []string{"X-EXAMPLE"},
		AuthConfig: &oauth2.Config{
			Endpoint: oauth2.Endpoint{
				AuthURL:  authServer.URL,
				TokenURL: authServer.URL,
			},
		},
		UnsafeState: true,
	})
	if err != nil {
		t.Fatalf("unable to init authenticator: %s", err)
	}

	var passedAuth bool
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedAuth = true
		w.WriteHeader(http.StatusOK)
	}))

	// hit once without any headers, expect a redirect/no pass
	r := httptest.NewRequest(http.MethodGet, "/xyz", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if passedAuth {
		t.Fatal("request passed the auth layer without including expected headers")
	}

	if got := w.Result(); got.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected to be redirected but got a status of %d instead",
			got.StatusCode)
	}

	// add headers, expect to get through
	r = httptest.NewRequest(http.MethodGet, "/xyz", nil)
	r.Header.Add("X-EXAMPLE", "1")

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !passedAuth {
		t.Fatal("request did not pass the auth layer despite including headers")
	}

	if got := w.Result(); got.StatusCode != http.StatusOK {
		t.Fatalf("expected to be get a 200 OK but got a status of %d instead",
			got.StatusCode)
	}
}

func setupAuthenticatorTest(t *testing.T) (*httptest.Server, *httptest.Server) {
	t.Helper()

	keyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(auth.JSONKeyResponse{
			Keys: []*auth.JSONKey{
				{
					Use: "sig",
					Kty: "RSA",
					Kid: "0b0bf186743471a1edcac3060d1256f9e4050ba8",
					N:   "0s9r8J5G5I77VpYWS-ttQ8GBDZBlxN_TZHl4DJHAi1WzvxQcP0hBPdASNqAnAuXA-ZxMpMtW_ovjhwo1Ncqpofd3c0H5mSzA9nsmmiex3AO7ZbkaGIdOcMYr4ttOFKZJn2giZWsfQuTlMEvcGyghViyy6l7t1-dMyxjbNOAVLVn25PHfWLbtffv-5EXFXt0Bp0wf0JjPghy4xXf3GjqqqaG_pOnmY_g2c6s8NwZG8dLymiqq0sta3URCUzDYnEHfx7Ol-grOYBOg6YjQP-gl0r5_uvB9Vl9jXKz-WcUUqVTuLp6S-CBstsOheUpSjX3vVP48KJIS4DX6NFHgjn8ooQ",
					E:   "AQAB",
				},
			},
		})
	}))
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
		  "access_token": "nah",
		  "expires_in": 3600,
		  "scope": "https://www.googleapis.com/auth/userinfo.email",
		  "token_type": "Bearer",
		  "id_token": "` + testAuthToken + `"
		}`))
	}))
	return keyServer, authServer
}
