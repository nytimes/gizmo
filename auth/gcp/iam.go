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
	iam "google.golang.org/api/iam/v1"
)

var (
	timeNow = func() time.Time { return time.Now() }

	// docs say up to 1 hour, this plays it safe?
	// https://cloud.google.com/compute/docs/instances/verifying-instance-identity#verify_signature
	defaultTokenTTL = time.Minute * 20
)

type iamKeySource struct {
	cf   func(context.Context) *http.Client
	name string
}

func NewIAMPublicKeySource(ctx context.Context, name string, clientFunc func(context.Context) *http.Client) (auth.PublicKeySource, error) {
	src := iamKeySource{cf: clientFunc, name: name}

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

	resp, err := svc.Projects.ServiceAccounts.Keys.List(s.name).Context(ctx).Do()
	if err != nil {
		return ks, errors.Wrap(err, "unable to list service account keys")
	}

	keys := map[string]*rsa.PublicKey{}
	for _, key := range resp.Keys {
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
			return ks, err
		}

		pkey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return ks, errors.Errorf("unexpected public key type: %T", cert.PublicKey)
		}

		_, name := path.Split(key.Name)
		keys[name] = pkey
	}

	return auth.NewPublicKeySet(keys, timeNow().Add(20*time.Minute)), nil
}

type IAMConfig struct {
	IAMAddress      string
	MetadataAddress string

	Audience            string
	Project             string
	ServiceAccountEmail string
}

// NewIAMTokenSource returns an oauth2.TokenSource that uses Google's IAM services
// to sign a JWT with the default service account and the given audience.
// Users should use the Identity token source if they can. This client is meant to be
// used as a bridge for users as they transition from the 1st generation App Engine
// runtime to the 2nd generation.
// This implementation can be used in the 2nd gen runtime as it can reuse an http.Client.
func NewIAMTokenSource(ctx context.Context, cfg IAMConfig) (oauth2.TokenSource, error) {
	tknSrc, err := defaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	svc, err := iam.New(oauth2.NewClient(ctx, tknSrc))
	if err != nil {
		return nil, err
	}
	src := &iamTokenSource{
		iamAddr:  cfg.IAMAddress,
		metaAddr: cfg.MetadataAddress,
		audience: cfg.Audience,

		svcAccount: cfg.ServiceAccountEmail,
		project:    cfg.Project,
		svc:        svc,
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
	src := &iamTokenSource{
		iamAddr:  cfg.IAMAddress,
		metaAddr: cfg.MetadataAddress,
		audience: cfg.Audience,

		svcAccount: cfg.ServiceAccountEmail,
		project:    cfg.Project,
	}

	tkn, err := src.ContextToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create initial token")
	}

	return &reuseTokenSource{t: tkn, new: src}, nil
}

type ContextTokenSource interface {
	ContextToken(context.Context) (*oauth2.Token, error)
}

type iamTokenSource struct {
	iamAddr  string
	metaAddr string
	audience string

	svcAccount, project string

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

	tkn, exp, err := s.newIAMToken(ctx, svc, s.audience)
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
	tkn, exp, err := s.newIAMToken(context.Background(), s.svc, s.audience)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: tkn,
		TokenType:   "Bearer",
		Expiry:      exp,
	}, nil
}

type IAMToken struct {
	Aud string `json:"aud"` // descriptor of the intended target.
	Exp int64  `json:"exp"` // the expiration time of the assertion (seconds since Unix epoch)
	Iat int64  `json:"iat"` // the time the assertion was issued (seconds since Unix epoch)

	Email string `json:"email"` // Email of the default service account
}

func (s iamTokenSource) newIAMToken(ctx context.Context, svc *iam.Service, audience string) (string, time.Time, error) {
	iss := timeNow()
	exp := iss.Add(defaultTokenTTL)
	payload, err := json.Marshal(IAMToken{
		Aud:   s.audience,
		Email: s.svcAccount,
		Exp:   exp.Unix(),
		Iat:   iss.Unix(),
	})
	if err != nil {
		return "", exp, errors.Wrap(err, "unable to encode JWT payload")
	}

	resp, err := svc.Projects.ServiceAccounts.SignJwt(
		fmt.Sprintf("projects/%s/serviceAccounts/%s",
			s.project, s.svcAccount),
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
