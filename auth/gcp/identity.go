package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/NYTimes/gizmo/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

type idKeySource struct {
	MetadataAddress string
	Issuers         map[string]struct{}
	CertURL         string

	hc *http.Client
}

func NewIdentityPublicKeySource(ctx context.Context) (auth.PublicKeySource, error) {
	hc := &http.Client{
		Timeout: 5 * time.Second,
	}
	src := idKeySource{
		hc: hc,
		Issuers: map[string]struct{}{
			"accounts.google.com":         struct{}{},
			"https://accounts.google.com": struct{}{},
		},
		CertURL: "https://www.googleapis.com/oauth2/v3/certs",
	}

	ks, err := src.Get(ctx)
	if err != nil {
		return nil, err
	}

	return auth.NewReusePublicKeySource(ks, src), nil
}

func (s idKeySource) Get(ctx context.Context) (auth.PublicKeySet, error) {
	return auth.NewPublicKeySetFromURL(s.hc, s.CertURL, time.Hour*2)
}

func NewIdentityTokenSource(audience string) (oauth2.TokenSource, error) {
	ts := &idTokenSource{
		audience: audience,
		mdc: metadata.NewClient(&http.Client{
			Timeout: 2 * time.Second,
		}),
	}
	tok, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return oauth2.ReuseTokenSource(tok, ts), nil
}

type idTokenSource struct {
	mdc      *metadata.Client
	audience string
}

// docs say up to 1 hour, this plays it safe?
// https://cloud.google.com/compute/docs/instances/verifying-instance-identity#verify_signature
var defaultTokenTTL = time.Minute * 20

func (c *idTokenSource) Token() (*oauth2.Token, error) {
	tkn, err := c.mdc.Get(
		fmt.Sprintf("instance/service-accounts/default/identity?audience=%s&format=full",
			c.audience))
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken: tkn,
		TokenType:   "Bearer",
		Expiry:      timeNow().Add(defaultTokenTTL),
	}, nil
}

type IdentityClaimSet struct {
	jws.ClaimSet

	// Email address of the default service account (only exists on GAE 2nd gen?)
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`

	// Google metadata info (appears to only exist on GCE?)
	Google map[string]interface{} `json:"google"`
}

func (s IdentityClaimSet) BaseClaims() *jws.ClaimSet {
	return &s.ClaimSet
}

func IdentityClaimsDecoderFunc(_ context.Context, b []byte) (auth.ClaimSetter, error) {
	var cs IdentityClaimSet
	err := json.Unmarshal(b, &cs)
	return cs, err
}

var timeNow = func() time.Time { return time.Now() }
