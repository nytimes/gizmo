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
// A function to set a new context is provided so users can cache/reuse tokens
// while dealing with the limitations of urlfetch and the 1st gen runtime.
func NewIAMTokenSource(ctx context.Context, cfg IAMConfig) (oauth2.TokenSource, func(context.Context) error, error) {
	src := &iamTokenSource{
		iamAddr:  cfg.IAMAddress,
		metaAddr: cfg.MetadataAddress,
		audience: cfg.Audience,

		svcAccount: cfg.ServiceAccountEmail,
		project:    cfg.Project,
	}

	err := src.setContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	tkn, err := src.Token()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create initial token")
	}

	return oauth2.ReuseTokenSource(tkn, src), src.setContext, nil
}

type iamTokenSource struct {
	iamAddr  string
	metaAddr string
	audience string

	svcAccount, project string

	// used just for the initial token fetch (1st gen GAE problems)
	ctx context.Context
	svc *iam.Service
	mu  sync.Mutex
}

var defaultTokenSource = google.DefaultTokenSource

func (s *iamTokenSource) setContext(ctx context.Context) error {
	tknSrc, err := defaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.svc, err = iam.New(oauth2.NewClient(ctx, tknSrc))
	if err != nil {
		return errors.Wrap(err, "unable to init iam client")
	}

	if s.iamAddr != "" {
		s.svc.BasePath = s.iamAddr
	}
	return nil
}

func (s iamTokenSource) Token() (*oauth2.Token, error) {
	tkn, exp, err := s.newIAMToken(s.ctx, s.audience)
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

func (s iamTokenSource) newIAMToken(ctx context.Context, audience string) (string, time.Time, error) {
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

	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.svc.Projects.ServiceAccounts.SignJwt(
		fmt.Sprintf("projects/%s/serviceAccounts/%s",
			s.project, s.svcAccount),
		&iam.SignJwtRequest{Payload: string(payload)}).Context(ctx).Do()
	if err != nil {
		return "", exp, errors.Wrap(err, "unable to sign JWT")
	}
	return resp.SignedJwt, exp, nil
}
