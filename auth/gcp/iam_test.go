package gcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	iam "google.golang.org/api/iam/v1"
)

func TestIAMKeySource(t *testing.T) {

	tests := []struct {
		name        string
		givenIAMErr bool

		wantErr bool
	}{
		{
			name: "normal success",
		},
		{
			name:        "iam error",
			givenIAMErr: true,

			wantErr: true,
		},
	}

	for _, test := range tests {
		const tokenValue = "iam-signed-jwt"
		t.Run(test.name, func(t *testing.T) {
			iamSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Log(r.URL.Path)
				if strings.Contains(r.URL.Path, "keys") {
					json.NewEncoder(w).Encode(iam.ListServiceAccountKeysResponse{
						Keys: []*iam.ServiceAccountKey{
							{Name: "8289d54280b76712de41cd2ef95972b123be9ac0"},
						},
					})
				} else {
					json.NewEncoder(w).Encode(iam.ServiceAccountKey{
						PublicKeyData: pubKey,
					})
				}
			}))
			if test.givenIAMErr {
				iamSvr.Close()
			} else {
				defer iamSvr.Close()
			}

			defaultTokenSource = func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
				return nil, nil
			}
			defer func() {
				defaultTokenSource = google.DefaultTokenSource
			}()

			cfg := IAMConfig{
				IAMAddress: iamSvr.URL,
			}

			ctx := context.Background()
			hc := func(_ context.Context) *http.Client {
				return http.DefaultClient
			}
			src, err := NewIAMPublicKeySource(ctx, cfg, hc)
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t but got %s", test.wantErr, err)
			}

			if src == nil {
				return
			}

			got, err := src.Get(ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t but got %s", test.wantErr, err)
			}

			if len(got.Keys) == 0 {
				t.Errorf("expected keys to be generated but got none")
			}
		})
	}
}

func TestIAMTokenSource(t *testing.T) {
	tests := []struct {
		name        string
		givenIAMErr bool

		wantErr bool
	}{
		{
			name: "normal success",
		},
		{
			name:        "iam error",
			givenIAMErr: true,

			wantErr: true,
		},
	}

	for _, test := range tests {
		const tokenValue = "iam-signed-jwt"
		t.Run(test.name, func(t *testing.T) {
			iamSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(iam.SignJwtResponse{
					SignedJwt: tokenValue,
				})
			}))
			if test.givenIAMErr {
				iamSvr.Close()
			} else {
				defer iamSvr.Close()
			}

			defaultTokenSource = func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
				return nil, nil
			}
			defer func() {
				defaultTokenSource = google.DefaultTokenSource
			}()

			cfg := IAMConfig{
				IAMAddress: iamSvr.URL,
			}

			ctx := context.Background()
			src, err := NewIAMTokenSource(ctx, cfg)
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t but got %s", test.wantErr, err)
			}

			if src == nil {
				return
			}

			got, err := src.Token()
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t but got %s", test.wantErr, err)
			}

			if got.AccessToken != tokenValue {
				t.Errorf("expected access token value of %s, got %s",
					tokenValue, got.AccessToken)
			}

			csrc, err := NewContextIAMTokenSource(ctx, cfg)
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t but got %s", test.wantErr, err)
			}

			if csrc == nil {
				return
			}

			got, err = csrc.ContextToken(ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("expected error? %t but got %s", test.wantErr, err)
			}

			if got.AccessToken != tokenValue {
				t.Errorf("expected access token value of %s, got %s",
					tokenValue, got.AccessToken)
			}
		})
	}

}

type testTokenSource struct{}

func (t testTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

const pubKey = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJakNDQWdxZ0F3SUJBZ0lJU3VCcXowQ0FMbk13RFFZSktvWklodmNOQVFFRkJRQXdOREV5TURBR0ExVUUKQXhNcGJubDBMV2RoYldWekxXUmxkaTVoY0hCemNHOTBMbWR6WlhKMmFXTmxZV05qYjNWdWRDNWpiMjB3SGhjTgpNVGd4TURJeU1EVXdNVEl3V2hjTk1UZ3hNVEEzTVRjeE5qSXdXakEwTVRJd01BWURWUVFERXlsdWVYUXRaMkZ0ClpYTXRaR1YyTG1Gd2NITndiM1F1WjNObGNuWnBZMlZoWTJOdmRXNTBMbU52YlRDQ0FTSXdEUVlKS29aSWh2Y04KQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCQUtQRmRZWkZUaUR1RmhWMDdKTHZuVjFhamZOL1hUdlZUbC9tUDNVbgpQVjdkakNVSTFRWFh2K2NhVmp4djVzOHc1WG9TWVpSOXpVYytlNXFpRHdndm5yQjMyeWkwajBhTEJXaHZiTmN3Ci9ZdFRBd1l3ZVhoVzM0ZzlsOFl5TkFHb0xJTDVVSy9ubXp3NVVDYm15V2pCZkdmYmdwNm5SRUI0dWhQWGM0MnoKZ1Y0cUJtS2pUclhvNk85OVVhbC9COU1rMFIzWWExMnBPclAyN0drZkxNZmFGSWlJSlhROW94aVdLMUw0U0pDdwpudDc0OEpoVFJHREtHWGtGZkxEMEFVTjAra2JsOU5hbXpVaVcvTmhXczJkd1FXOGN2YUxuWG11NENraEM1aVo1CmU4RHJvSEswMDFCd21HNmhrU1ZxN2laNnFZU2J6ekwrc3NpUTB5KzVwZDBkcXZrQ0F3RUFBYU00TURZd0RBWUQKVlIwVEFRSC9CQUl3QURBT0JnTlZIUThCQWY4RUJBTUNCNEF3RmdZRFZSMGxBUUgvQkF3d0NnWUlLd1lCQlFVSApBd0l3RFFZSktvWklodmNOQVFFRkJRQURnZ0VCQUk2eHExTzloRm4wN1lDbzhJV3E1Mk1PZVFnenN5YXBKbXVHCmhaUWF3Q1l1SmMwSFo3RXJRYkxSejdYRDBNbzRwQlFHSGl2SGtuMW1GcUxna2c5eHdISlhHVnQzbC9RWlczeWQKMDhXa0RvMlhjbEkwaE1pT3gxSHBtREVsT1FheE5Gd2NlV1VlN09hck5Da0dHVGsySEZsK3QzQkxWcDVYWnFEaQpiQlZpY0tSTTREczF6dURtQzhtUWppbjVxR0VYM1IrZ2hacGhDcnRjdC9yTWF6eW5iUHdDbnFzRDZVMFNRMXZ6CjF2RnBPRDd4cnlVZ0VuZTh4SnlEVy9aeHkzTXBIeThMTWtiSjZjVEZjaGVzOFlvaTVFYkVyYTk2NEU4SVg4dlcKdHgwU2VUdmhoT0ltQ2VsdE9UTkZmMmkxSitIY2Y0V3JaNzN4cFQ4YnV3SjZyeEkwdDRnPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
