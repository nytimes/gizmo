package gcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

func TestVerifyIdentity(t *testing.T) {
	tests := []struct {
		name string

		givenEmails   []string
		givenAudience string

		givenClaims interface{}

		wantVerified bool
	}{
		{
			name: "normal success",

			givenEmails:   []string{"jp@example.com"},
			givenAudience: "example.com",

			givenClaims: IdentityClaimSet{
				ClaimSet: jws.ClaimSet{
					Aud: "example.com",
					Iss: "https://accounts.google.com",
				},
				Email:         "jp@example.com",
				EmailVerified: true,
			},

			wantVerified: true,
		},
		{
			name: "invalid issuer",

			givenEmails:   []string{"jp@example.com"},
			givenAudience: "example.com",

			givenClaims: IdentityClaimSet{
				ClaimSet: jws.ClaimSet{
					Aud: "example.com",
					Iss: "https://google.com",
				},
				Email: "jp@example.com",
			},

			wantVerified: false,
		},
		{
			name: "unverified email",

			givenEmails:   []string{"jp@example.com"},
			givenAudience: "example.com",

			givenClaims: IdentityClaimSet{
				ClaimSet: jws.ClaimSet{
					Aud: "example.com",
					Iss: "https://accounts.google.com",
				},
				Email: "jp@example.com",
			},

			wantVerified: false,
		},
		{
			name: "invalid claims type",

			givenEmails:   []string{"jp@example.com"},
			givenAudience: "example.com",

			givenClaims: jws.ClaimSet{
				Aud: "google.com",
				Iss: "https://accounts.google.com",
			},

			wantVerified: false,
		},
		{
			name: "invalid audience",

			givenEmails:   []string{"jp@example.com"},
			givenAudience: "example.com",

			givenClaims: IdentityClaimSet{
				ClaimSet: jws.ClaimSet{
					Aud: "google.com",
					Iss: "https://accounts.google.com",
				},
				Email:         "jp@example.com",
				EmailVerified: true,
			},

			wantVerified: false,
		},
		{
			name: "bad email",

			givenEmails:   []string{"jape@example.com"},
			givenAudience: "example.com",

			givenClaims: IdentityClaimSet{
				ClaimSet: jws.ClaimSet{
					Aud: "example.com",
					Iss: "https://accounts.google.com",
				},
				Email:         "jp@example.com",
				EmailVerified: true,
			},

			wantVerified: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vf := VerifyIdentityEmails(context.Background(), test.givenEmails, test.givenAudience)

			got := vf(context.Background(), test.givenClaims)
			if got != test.wantVerified {
				t.Errorf("expected verfied? %t, got %t", test.wantVerified, got)
			}
		})
	}

}

func TestIdentityTokenSource(t *testing.T) {
	tests := []struct {
		name string

		givenMetaStatus int
		givenMetaErr    bool

		wantErr bool
	}{
		{
			name:            "success",
			givenMetaStatus: http.StatusOK,
		},
		{
			name:            "bad status from metadata",
			givenMetaStatus: http.StatusNotFound,

			wantErr: true,
		},
		{
			name:         "bad connection to metadata",
			givenMetaErr: true,

			wantErr: true,
		},
	}

	for _, test := range tests {

		testTokenString := "blargityblarg"
		testToken := oauth2.Token{
			AccessToken: testTokenString,
			TokenType:   "Bearer",
			Expiry:      timeNow().Add(20 * time.Minute),
		}
		t.Run(test.name, func(t *testing.T) {
			metaSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.String(), "audience=my-aud&format") {
					t.Errorf("expected 'my-aud' audience in metadata call but did not see it: %s",
						r.URL.String())
				}
				w.WriteHeader(test.givenMetaStatus)
				w.Write([]byte(testTokenString))
			}))
			defer metaSvr.Close()
			if test.givenMetaErr {
				metaSvr.Close()
			}

			cfg := IdentityConfig{
				MetadataAddress: metaSvr.URL + "/",
				Audience:        "my-aud",
			}

			src, err := NewIdentityTokenSource(cfg)

			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t, but got  %s", test.wantErr, err)
			}

			if src == nil {
				return
			}

			got, err := src.Token()
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t, but got  %s", test.wantErr, err)
				return
			}

			if cmp.Equal(got, testToken) {
				t.Errorf("unpexpected token returned: %s", cmp.Diff(got, testToken))
			}
		})
	}
}
