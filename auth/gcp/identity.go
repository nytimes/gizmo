package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NYTimes/gizmo/auth"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

// IdentityConfig contains the information required for generating or verifying identity
// JWTs.
type IdentityConfig struct {
	Audience string `envconfig:"ID_AUDIENCE"`

	CertURL string `envconfig:"ID_CERT_URL"` // optional override for public key source

	Client *http.Client // optional override

	MetadataAddress string `envconfig:"ID_METADATA_ADDR"` // optional override for token and email retrieval
}

type idKeySource struct {
	cfg IdentityConfig
}

// NewDefaultIdentityVerifier will verify tokens that have the same default service
// account as the server running this verifier.
func NewDefaultIdentityVerifier(ctx context.Context, cfg IdentityConfig) (*auth.Verifier, error) {
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 2 * time.Second}
	}

	ks, err := NewIdentityPublicKeySource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	eml, err := GetDefaultEmail(ctx, "", cfg.Client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get default email")
	}

	return auth.NewVerifier(ks,
		IdentityClaimsDecoderFunc,
		VerifyIdentityEmails(ctx, []string{eml}, cfg.Audience)), nil
}

// NewIdentityPublicKeySource fetches Google's public oauth2 certificates to be used with
// the auth.Verifier tool.
func NewIdentityPublicKeySource(ctx context.Context, cfg IdentityConfig) (auth.PublicKeySource, error) {
	if cfg.Client == nil {
		cfg.Client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	if cfg.CertURL == "" {
		cfg.CertURL = "https://www.googleapis.com/oauth2/v3/certs"
	}

	src := idKeySource{cfg: cfg}

	ks, err := src.Get(ctx)
	if err != nil {
		return nil, err
	}

	return auth.NewReusePublicKeySource(ks, src), nil
}

func (s idKeySource) Get(ctx context.Context) (auth.PublicKeySet, error) {
	return auth.NewPublicKeySetFromURL(s.cfg.Client, s.cfg.CertURL, time.Hour*2)
}

// NewIdentityTokenSource will use the GCP metadata services to generate GCP Identity
// tokens. More information on asserting GCP identities can be found here:
// https://cloud.google.com/compute/docs/instances/verifying-instance-identity
func NewIdentityTokenSource(cfg IdentityConfig) (oauth2.TokenSource, error) {
	if cfg.Client == nil {
		cfg.Client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	if cfg.MetadataAddress == "" {
		cfg.MetadataAddress = "http://metadata/computeMetadata/v1/"
	}

	ts := &idTokenSource{cfg: cfg}

	tok, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return oauth2.ReuseTokenSource(tok, ts), nil
}

type idTokenSource struct {
	cfg IdentityConfig
}

func (c *idTokenSource) Token() (*oauth2.Token, error) {
	suffix := fmt.Sprintf("instance/service-accounts/default/identity?audience=%s&format=full",
		c.cfg.Audience)

	tkn, err := metadataGet(context.Background(), c.cfg.MetadataAddress, c.cfg.Client, suffix)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get token")
	}

	return &oauth2.Token{
		AccessToken: tkn,
		TokenType:   "Bearer",
		Expiry:      timeNow().Add(defaultTokenTTL),
	}, nil
}

// IdentityClaimSet holds all the expected values for the various versions of the GCP
// identity token.
// More details:
// https://cloud.google.com/compute/docs/instances/verifying-instance-identity#payload
// https://developers.google.com/identity/sign-in/web/backend-auth#calling-the-tokeninfo-endpoint
type IdentityClaimSet struct {
	jws.ClaimSet

	// Email address of the default service account (only exists on GAE 2nd gen?)
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`

	// Google metadata info (appears to only exist on GCE?)
	Google map[string]interface{} `json:"google"`
}

// BaseClaims implements the auth.ClaimSetter interface.
func (s IdentityClaimSet) BaseClaims() *jws.ClaimSet {
	return &s.ClaimSet
}

// IdentityClaimsDecoderFunc is an auth.ClaimsDecoderFunc for GCP identity tokens.
func IdentityClaimsDecoderFunc(_ context.Context, b []byte) (auth.ClaimSetter, error) {
	var cs IdentityClaimSet
	err := json.Unmarshal(b, &cs)
	if err != nil {
		return cs, errors.Wrap(auth.ErrBadCreds, err.Error())
	}
	return cs, nil
}

// IdentityVerifyFunc auth.VerifyFunc wrapper around the IdentityClaimSet.
func IdentityVerifyFunc(vf func(ctx context.Context, cs IdentityClaimSet) bool) auth.VerifyFunc {
	return func(ctx context.Context, c interface{}) bool {
		ics, ok := c.(IdentityClaimSet)
		if !ok {
			return false
		}
		return vf(ctx, ics)
	}
}

// Issuers contains the known Google account issuers for identity tokens.
var Issuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

// ValidIdentityClaims ensures the token audience and issuers match expectations.
func ValidIdentityClaims(cs IdentityClaimSet, audience string) bool {
	if cs.Aud != audience {
		return false
	}
	return Issuers[cs.Iss]
}

// VerifyIdentityEmails is an auth.VerifyFunc that ensures IdentityClaimSets are valid
// and have the expected email and audience in their payload.
func VerifyIdentityEmails(ctx context.Context, emails []string, audience string) auth.VerifyFunc {
	emls := map[string]bool{}
	for _, e := range emails {
		emls[e] = true
	}
	return IdentityVerifyFunc(func(ctx context.Context, cs IdentityClaimSet) bool {
		if !ValidIdentityClaims(cs, audience) {
			return false
		}
		if !cs.EmailVerified {
			return false
		}
		return emls[cs.Email]
	})
}
