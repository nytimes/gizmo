package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2/jws"
)

func TestVerifyRequest(t *testing.T) {
	tests := []struct {
		name string

		givenBadToken   bool
		givenBadRequest bool

		wantVerified bool
		wantErr      bool
	}{
		{
			name: "normal route, success",

			wantVerified: true,
		},
		{
			name: "bad token",

			givenBadToken: true,

			wantErr: true,
		},
		{
			name: "bad request",

			givenBadRequest: true,

			wantErr: true,
		},
	}

	prv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("unable to generate private key: %s", err)
	}

	keyID := "the-key"
	testTime := time.Date(2018, 10, 29, 12, 0, 0, 0, time.UTC)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TimeNow = func() time.Time { return testTime }

			token, err := encode(
				&jws.Header{Algorithm: "RS256", Typ: "JWT", KeyID: keyID},
				testClaims{
					ClaimSet: jws.ClaimSet{
						Iss: "example.com",
						Iat: testTime.Add(-5 * time.Second).Unix(),
						Exp: testTime.Add(5 * time.Second).Unix(),
						Aud: "example.com",
					},
				}, prv)
			if err != nil {
				t.Fatalf("unable to encode token: %s", err)
			}

			decoderFunc := func(_ context.Context, b []byte) (ClaimSetter, error) {
				var c testClaims
				err := json.Unmarshal(b, &c)
				return c, err
			}

			verifyFunc := func(_ context.Context, c interface{}) bool {
				return true
			}

			ks := testKeySource{
				keys: PublicKeySet{
					Expiry: TimeNow().Add(time.Hour),
					Keys: map[string]*rsa.PublicKey{
						keyID: &prv.PublicKey,
					},
				},
			}

			vrfy := NewVerifier(ks, decoderFunc, verifyFunc)

			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			token = "Bearer " + token
			if test.givenBadToken {
				token = "ASDFLKANSDFLKJ"
			}

			if !test.givenBadRequest {
				r.Header.Add("Authorization", token)
			}

			verified, err := vrfy.VerifyRequest(r)
			if (err != nil) != test.wantErr {
				t.Errorf("unexpected error? %t, got %s", test.wantErr, err)
			}

			if verified != test.wantVerified {
				t.Errorf("wanted verified? %t, got %t", test.wantVerified, verified)
			}
		})
	}

}

func TestVerifyInboundKit(t *testing.T) {
	tests := []struct {
		name string

		givenBadToken   bool
		givenBadContext bool

		wantVerified bool
		wantErr      bool
	}{
		{
			name: "normal route, success",

			wantVerified: true,
		},
		{
			name: "bad token",

			givenBadToken: true,

			wantErr: true,
		},
		{
			name: "bad context",

			givenBadContext: true,

			wantErr: true,
		},
	}

	prv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("unable to generate private key: %s", err)
	}

	keyID := "the-key"
	testTime := time.Date(2018, 10, 29, 12, 0, 0, 0, time.UTC)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TimeNow = func() time.Time { return testTime }

			token, err := encode(
				&jws.Header{Algorithm: "RS256", Typ: "JWT", KeyID: keyID},
				testClaims{
					ClaimSet: jws.ClaimSet{
						Iss: "example.com",
						Iat: testTime.Add(-5 * time.Second).Unix(),
						Exp: testTime.Add(5 * time.Second).Unix(),
						Aud: "example.com",
					},
				}, prv)
			if err != nil {
				t.Fatalf("unable to encode token: %s", err)
			}

			decoderFunc := func(_ context.Context, b []byte) (ClaimSetter, error) {
				var c testClaims
				err := json.Unmarshal(b, &c)
				return c, err
			}

			verifyFunc := func(_ context.Context, c interface{}) bool {
				return true
			}

			ks := testKeySource{
				keys: PublicKeySet{
					Expiry: TimeNow().Add(time.Hour),
					Keys: map[string]*rsa.PublicKey{
						keyID: &prv.PublicKey,
					},
				},
			}

			vrfy := NewVerifier(ks, decoderFunc, verifyFunc)

			ctx := context.Background()

			token = "Bearer " + token
			if test.givenBadToken {
				token = "ASDFLKANSDFLKJ"
			}

			if !test.givenBadContext {
				ctx = context.WithValue(ctx, httptransport.ContextKeyRequestAuthorization, token)
			}

			verified, err := vrfy.VerifyInboundKitContext(ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("unexpected error? %t, got %s", test.wantErr, err)
			}

			if verified != test.wantVerified {
				t.Errorf("wanted verified? %t, got %t", test.wantVerified, verified)
			}
		})
	}

}

func TestVerify(t *testing.T) {
	testTime := time.Date(2018, 10, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string

		givenClaims   testClaims
		givenBadToken bool
		givenTime     time.Time
		givenKeyError error
		givenBadKeyID bool

		wantVerified bool
		wantErr      bool
	}{
		{
			name: "normal JWT, success",
			givenClaims: testClaims{
				ClaimSet: jws.ClaimSet{
					Iss: "example.com",
					Iat: testTime.Add(-5 * time.Second).Unix(),
					Exp: testTime.Add(5 * time.Second).Unix(),
					Aud: "example.com",
				},
			},
			givenTime: testTime,

			wantVerified: true,
		},
		{
			name: "invalid Iat",
			givenClaims: testClaims{
				ClaimSet: jws.ClaimSet{
					Iss: "example.com",
					Iat: testTime.Add(10 * time.Minute).Unix(),
					Exp: testTime.Add(5 * time.Second).Unix(),
					Aud: "example.com",
				},
			},
			givenTime: testTime,

			wantErr: true,
		},
		{
			name: "expired claims",
			givenClaims: testClaims{
				ClaimSet: jws.ClaimSet{
					Iss: "example.com",
					Iat: testTime.Add(-15 * time.Minute).Unix(),
					Exp: testTime.Add(-10 * time.Minute).Unix(),
					Aud: "example.com",
				},
			},
			givenTime: testTime,

			wantErr: true,
		},
		{
			name: "unable to get keys",
			givenClaims: testClaims{
				ClaimSet: jws.ClaimSet{
					Iss: "example.com",
					Iat: testTime.Add(-15 * time.Minute).Unix(),
					Exp: testTime.Add(-10 * time.Minute).Unix(),
					Aud: "example.com",
				},
			},
			givenTime:     testTime,
			givenKeyError: errors.New("angry computer"),

			wantErr: true,
		},
		{
			name: "invalid key id",
			givenClaims: testClaims{
				ClaimSet: jws.ClaimSet{
					Iss: "example.com",
					Iat: testTime.Add(-15 * time.Minute).Unix(),
					Exp: testTime.Add(-10 * time.Minute).Unix(),
					Aud: "example.com",
				},
			},
			givenTime:     testTime,
			givenBadKeyID: true,

			wantErr: true,
		},
		{
			name: "invalid token",
			givenClaims: testClaims{
				ClaimSet: jws.ClaimSet{
					Iss: "example.com",
					Iat: testTime.Add(-15 * time.Minute).Unix(),
					Exp: testTime.Add(-10 * time.Minute).Unix(),
					Aud: "example.com",
				},
			},
			givenTime:     testTime,
			givenBadToken: true,

			wantErr: true,
		},
	}

	prv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("unable to generate private key: %s", err)
	}

	keyID := "the-key"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TimeNow = func() time.Time { return testTime }

			token, err := encode(
				&jws.Header{Algorithm: "RS256", Typ: "JWT", KeyID: keyID},
				test.givenClaims, prv)
			if err != nil {
				t.Fatalf("unable to encode token: %s", err)
			}

			decoderFunc := func(_ context.Context, b []byte) (ClaimSetter, error) {
				var c testClaims
				err := json.Unmarshal(b, &c)
				return c, err
			}

			verifyFunc := func(_ context.Context, c interface{}) bool {
				tc, ok := c.(testClaims)
				if !ok {
					t.Errorf("expected testClaims type, got %T", c)
				}
				if !cmp.Equal(tc, test.givenClaims) {
					t.Errorf("claims were not what we expected: %s", cmp.Diff(tc, test.givenClaims))
				}
				return true
			}

			kid := keyID
			if test.givenBadKeyID {
				kid = "blah"
			}

			ks := testKeySource{
				keys: PublicKeySet{
					Expiry: TimeNow().Add(time.Hour),
					Keys: map[string]*rsa.PublicKey{
						kid: &prv.PublicKey,
					},
				},
				err: test.givenKeyError,
			}

			vrfy := NewVerifier(ks, decoderFunc, verifyFunc)

			if test.givenBadToken {
				token = "ASDFLKANSDFLKJ"
			}

			verified, err := vrfy.Verify(context.Background(), token)
			if (err != nil) != test.wantErr {
				t.Errorf("unexpected error? %t, got %s", test.wantErr, err)
			}

			if verified != test.wantVerified {
				t.Errorf("wanted verified? %t, got %t", test.wantVerified, verified)
			}
		})
	}

}

// taken from jws package bc we just have an interface type for signing
func encodeWithSigner(header *jws.Header, c interface{}, sg jws.Signer) (string, error) {
	bh, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	head := base64.RawURLEncoding.EncodeToString(bh)

	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	cs := base64.RawURLEncoding.EncodeToString(b)
	ss := fmt.Sprintf("%s.%s", head, cs)
	sig, err := sg([]byte(ss))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", ss, base64.RawURLEncoding.EncodeToString(sig)), nil
}

func encode(header *jws.Header, c interface{}, key *rsa.PrivateKey) (string, error) {
	sg := func(data []byte) (sig []byte, err error) {
		h := sha256.New()
		h.Write(data)
		return rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h.Sum(nil))
	}
	return encodeWithSigner(header, c, sg)
}

type testClaims struct {
	jws.ClaimSet
}

func (t testClaims) BaseClaims() *jws.ClaimSet {
	return &t.ClaimSet
}
