package gcp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/NYTimes/gizmo/auth"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
	iam "google.golang.org/api/iam/v1"
)

var (
	timeNow = func() time.Time { return time.Now() }

	// docs say up to 1 hour, this plays it safe?
	// https://cloud.google.com/compute/docs/instances/verifying-instance-identity#verify_signature
	defaultTokenTTL = time.Minute * 20
)

// IAMClaimSet contains just an email for service account identification.
type IAMClaimSet struct {
	jws.ClaimSet

	// Email address of the default service account
	Email string `json:"email"`
}

// NewDefaultIAMVerifier will verify tokens that have the same default service account as
// the server running this verifier.
func NewDefaultIAMVerifier(ctx context.Context, cfg IAMConfig, clientFunc func(context.Context) *http.Client) (*auth.Verifier, error) {
	var err error
	if cfg.ServiceAccountEmail == "" {
		cfg.ServiceAccountEmail, err = GetDefaultEmail(ctx, "", clientFunc(ctx))
		if err != nil {
			return nil, errors.Wrap(err, "unable to get default email")
		}
	}

	ks, err := NewIAMPublicKeySource(ctx, cfg, clientFunc)
	if err != nil {
		return nil, err
	}

	return auth.NewVerifier(ks,
		IAMClaimsDecoderFunc,
		VerifyIAMEmails(ctx, []string{cfg.ServiceAccountEmail}, cfg.Audience)), nil
}

// BaseClaims implements the auth.ClaimSetter interface.
func (s IAMClaimSet) BaseClaims() *jws.ClaimSet {
	return &s.ClaimSet
}

// IAMClaimsDecoderFunc is an auth.ClaimsDecoderFunc for GCP identity tokens.
func IAMClaimsDecoderFunc(_ context.Context, b []byte) (auth.ClaimSetter, error) {
	var cs IAMClaimSet
	err := json.Unmarshal(b, &cs)
	return cs, err
}

// IAMVerifyFunc auth.VerifyFunc wrapper around the IAMClaimSet.
func IAMVerifyFunc(vf func(ctx context.Context, cs IAMClaimSet) bool) auth.VerifyFunc {
	return func(ctx context.Context, c interface{}) bool {
		ics, ok := c.(IAMClaimSet)
		if !ok {
			return false
		}
		return vf(ctx, ics)
	}
}

// ValidIAMClaims ensures the token audience issuers matches expectations.
func ValidIAMClaims(cs IAMClaimSet, audience string) bool {
	return cs.Aud == audience
}

// VerifyIAMEmails is an auth.VerifyFunc that ensures IAMClaimSets are valid
// and have the expected email and audience in their payload.
func VerifyIAMEmails(ctx context.Context, emails []string, audience string) auth.VerifyFunc {
	emls := map[string]bool{}
	for _, e := range emails {
		emls[e] = true
	}
	return IAMVerifyFunc(func(ctx context.Context, cs IAMClaimSet) bool {
		if !ValidIAMClaims(cs, audience) {
			return false
		}
		return emls[cs.Email]
	})
}

type iamKeySource struct {
	cf  func(context.Context) *http.Client
	cfg IAMConfig
}

// NewIAMPublicKeySource returns a PublicKeySource that uses the Google IAM service
// for fetching public keys of a given service account. The function for returning an
// HTTP client is to allow 1st generation App Engine users to lean on urlfetch.
func NewIAMPublicKeySource(ctx context.Context, cfg IAMConfig, clientFunc func(context.Context) *http.Client) (auth.PublicKeySource, error) {
	src := iamKeySource{cf: clientFunc, cfg: cfg}

	ks, err := src.Get(ctx)
	if err != nil {
		return nil, err
	}

	return auth.NewReusePublicKeySource(ks, src), nil
}

func (s iamKeySource) Get(ctx context.Context) (auth.PublicKeySet, error) {
	var ks auth.PublicKeySet

	// for the sake of GAE standard users who have to use a different *http.Client on
	// each request, we're going to init a new iam.Service on each fetch.
	// since this is cached, it should hopefully not be a huge issue
	svc, err := iam.New(s.cf(ctx))
	if err != nil {
		return ks, errors.Wrap(err, "unable to init iam client")
	}

	if s.cfg.IAMAddress != "" {
		svc.BasePath = s.cfg.IAMAddress
	}

	name := fmt.Sprintf("projects/%s/serviceAccounts/%s",
		s.cfg.Project, s.cfg.ServiceAccountEmail)
	resp, err := svc.Projects.ServiceAccounts.Keys.List(name).Context(ctx).Do()
	if err != nil {
		return ks, errors.Wrap(err, "unable to list service account keys")
	}

	keys := map[string]*rsa.PublicKey{}
	for _, keyData := range resp.Keys {
		// we need to fetch each key's PublicKey data since List only returns metadata.
		key, err := svc.Projects.ServiceAccounts.Keys.Get(keyData.Name).
			PublicKeyType("TYPE_X509_PEM_FILE").Context(ctx).Do()
		if err != nil {
			return ks, errors.Wrap(err, "unable to get public key data")
		}

		pemBytes, err := base64.StdEncoding.DecodeString(key.PublicKeyData)
		if err != nil {
			return ks, err
		}

		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return ks, errors.New("Unable to find pem block in key")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return ks, errors.Wrap(err, "unable to parse x509 certificate")
		}

		pkey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return ks, errors.Errorf("unexpected public key type: %T", cert.PublicKey)
		}

		_, name := path.Split(key.Name)
		keys[name] = pkey
	}

	return auth.PublicKeySet{Keys: keys, Expiry: timeNow().Add(20 * time.Minute)}, nil
}

// IAMConfig contains the information required for generating or verifying IAM JWTs.
type IAMConfig struct {
	IAMAddress string `envconfig:"IAM_ADDR"` // optional, for testing

	Audience            string `envconfig:"IAM_AUDIENCE"`
	Project             string `envconfig:"IAM_PROJECT"`
	ServiceAccountEmail string `envconfig:"IAM_SERVICE_ACCOUNT_EMAIL"`

	// JSON contains the raw bytes from a JSON credentials file.
	// This field may be nil if authentication is provided by the
	// environment and not with a credentials file, e.g. when code is
	// running on Google Cloud Platform.
	JSON []byte
}

// NewIAMTokenSource returns an oauth2.TokenSource that uses Google's IAM services
// to sign a JWT with the default service account and the given audience.
// Users should use the Identity token source if they can. This client is meant to be
// used as a bridge for users as they transition from the 1st generation App Engine
// runtime to the 2nd generation.
// This implementation can be used in the 2nd gen runtime as it can reuse an http.Client.
func NewIAMTokenSource(ctx context.Context, cfg IAMConfig) (oauth2.TokenSource, error) {
	var (
		err    error
		tknSrc oauth2.TokenSource
	)
	if cfg.JSON != nil {
		creds, err := google.CredentialsFromJSON(ctx, cfg.JSON, iam.CloudPlatformScope)
		if err != nil {
			return nil, err
		}
		tknSrc = creds.TokenSource
	} else {
		tknSrc, err = defaultTokenSource(ctx, iam.CloudPlatformScope)
	}
	if err != nil {
		return nil, err
	}

	svc, err := iam.New(oauth2.NewClient(ctx, tknSrc))
	if err != nil {
		return nil, err
	}

	if cfg.IAMAddress != "" {
		svc.BasePath = cfg.IAMAddress
	}

	src := &iamTokenSource{
		cfg: cfg,
		svc: svc,
	}

	tkn, err := src.Token()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create initial token")
	}

	return oauth2.ReuseTokenSource(tkn, src), nil
}

// NewContextIAMTokenSource returns an oauth2.TokenSource that uses Google's IAM services
// to sign a JWT with the default service account and the given audience.
// Users should use the Identity token source if they can. This client is meant to be
// used as a bridge for users as they transition from the 1st generation App Engine
// runtime to the 2nd generation.
// This implementation can be used in the 1st gen runtime as it allows users to pass a
// context.Context while fetching the token. The context allows the implementation to
// reuse clients while changing out the HTTP client under the hood.
func NewContextIAMTokenSource(ctx context.Context, cfg IAMConfig) (ContextTokenSource, error) {
	src := &iamTokenSource{cfg: cfg}

	tkn, err := src.ContextToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create initial token")
	}

	return &reuseTokenSource{t: tkn, new: src}, nil
}

// ContextTokenSource is an oauth2.TokenSource that is capable of running on the 1st
// generation App Engine environment because it can create a urlfetch.Client from the
// given context.
type ContextTokenSource interface {
	ContextToken(context.Context) (*oauth2.Token, error)
}

type iamTokenSource struct {
	cfg IAMConfig

	svc *iam.Service
}

var defaultTokenSource = google.DefaultTokenSource

func (s iamTokenSource) ContextToken(ctx context.Context) (*oauth2.Token, error) {
	tknSrc, err := defaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	svc, err := iam.New(oauth2.NewClient(ctx, tknSrc))
	if err != nil {
		return nil, err
	}

	if s.cfg.IAMAddress != "" {
		svc.BasePath = s.cfg.IAMAddress
	}

	tkn, exp, err := s.newIAMToken(ctx, svc)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: tkn,
		TokenType:   "Bearer",
		Expiry:      exp,
	}, nil
}

func (s iamTokenSource) Token() (*oauth2.Token, error) {
	tkn, exp, err := s.newIAMToken(context.Background(), s.svc)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: tkn,
		TokenType:   "Bearer",
		Expiry:      exp,
	}, nil
}

func (s iamTokenSource) newIAMToken(ctx context.Context, svc *iam.Service) (string, time.Time, error) {
	iss := timeNow()
	exp := iss.Add(defaultTokenTTL)
	payload, err := json.Marshal(IAMClaimSet{
		ClaimSet: jws.ClaimSet{
			Iss: s.cfg.ServiceAccountEmail,
			Sub: s.cfg.ServiceAccountEmail,
			Aud: s.cfg.Audience,
			Exp: exp.Unix(),
			Iat: iss.Unix(),
		},
		Email: s.cfg.ServiceAccountEmail,
	})
	if err != nil {
		return "", exp, errors.Wrap(err, "unable to encode JWT payload")
	}

	resp, err := svc.Projects.ServiceAccounts.SignJwt(
		fmt.Sprintf("projects/%s/serviceAccounts/%s",
			s.cfg.Project, s.cfg.ServiceAccountEmail),
		&iam.SignJwtRequest{Payload: string(payload)}).Context(ctx).Do()
	if err != nil {
		return "", exp, errors.Wrap(err, "unable to sign JWT")
	}
	return resp.SignedJwt, exp, nil
}

// TAKEN FROM golang.org/x/oauth2 so we can add context bc GAE 1st gen + urlfetch.
// reuseCtxTokenSource is a TokenSource that holds a single token in memory
// and validates its expiry before each call to retrieve it with
// Token. If it's expired, it will be auto-refreshed using the
// new TokenSource.
type reuseTokenSource struct {
	new ContextTokenSource // called when t is expired.

	mu sync.Mutex // guards t
	t  *oauth2.Token
}

// Token returns the current token if it's still valid, else will
// refresh the current token (using r.Context for HTTP client
// information) and return the new one.
func (s *reuseTokenSource) ContextToken(ctx context.Context) (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t.Valid() {
		return s.t, nil
	}
	t, err := s.new.ContextToken(ctx)
	if err != nil {
		return nil, err
	}
	s.t = t
	return t, nil
}
