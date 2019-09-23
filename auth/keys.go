package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// PublicKeySource is to be used by servers who need to acquire public key sets for
// verifying inbound request's JWTs.
type PublicKeySource interface {
	Get(context.Context) (PublicKeySet, error)
}

// NewReusePublicKeySource is a wrapper around PublicKeySources to only fetch a new key
// set once the current key cache has expired.
func NewReusePublicKeySource(ks PublicKeySet, src PublicKeySource) PublicKeySource {
	return &reuseKeySource{ks: ks, src: src}
}

type reuseKeySource struct {
	src PublicKeySource

	mu sync.Mutex
	ks PublicKeySet
}

func (r *reuseKeySource) Get(ctx context.Context) (PublicKeySet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ks.Expired() {
		var err error
		r.ks, err = r.src.Get(ctx)
		return r.ks, err
	}
	return r.ks, nil
}

// PublicKeySet contains a set of keys acquired from a JWKS that has an expiration.
type PublicKeySet struct {
	Expiry time.Time
	Keys   map[string]*rsa.PublicKey
}

// Expired will return true if the current key set is expire according to its Expiry
// field.
func (ks PublicKeySet) Expired() bool {
	return TimeNow().After(ks.Expiry)
}

// GetKey will look for the given key ID in the key set and return it, if it exists.
func (ks PublicKeySet) GetKey(id string) (*rsa.PublicKey, error) {
	if len(ks.Keys) == 0 {
		return nil, errors.New("no public keys found")
	}
	key, ok := ks.Keys[id]
	if !ok {
		return nil, errors.Wrapf(ErrBadCreds, "key [%s] not found in set of size %d", id, len(ks.Keys))
	}
	return key, nil
}

// JSONKey represents a public or private key in JWK format.
type JSONKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JSONKeyResponse represents a JWK Set object.
type JSONKeyResponse struct {
	Keys []*JSONKey `json:"keys"`
}

var reMaxAge = regexp.MustCompile("max-age=([0-9]*)")

// NewPublicKeySetFromURL will attempt to fetch a JWKS from the given URL and parse it
// into a PublicKeySet. The endpoint the URL points to must return the same format as the
// JSONKeyResponse struct.
func NewPublicKeySetFromURL(hc *http.Client, url string, defaultTTL time.Duration) (PublicKeySet, error) {
	var ks PublicKeySet
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ks, errors.Wrap(err, "unable to create request")
	}

	resp, err := hc.Do(r)
	if err != nil {
		return ks, err
	}
	defer resp.Body.Close()

	ttl := defaultTTL
	if ccHeader := resp.Header.Get("cache-control"); ccHeader != "" {
		if match := reMaxAge.FindStringSubmatch(ccHeader); len(match) > 1 {
			maxAgeSeconds, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				return ks, errors.Wrap(err, "unable to parse cache-control max age")
			}
			ttl = time.Second * time.Duration(maxAgeSeconds)
		}
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ks, errors.Wrap(err, "unable to read response")
	}

	return NewPublicKeySetFromJSON(payload, ttl)
}

// NewPublicKeySetFromJSON will accept a JSON payload in the format of the
// JSONKeyResponse and parse it into a PublicKeySet.
func NewPublicKeySetFromJSON(payload []byte, ttl time.Duration) (PublicKeySet, error) {
	var (
		ks   PublicKeySet
		keys JSONKeyResponse
	)
	err := json.Unmarshal(payload, &keys)
	if err != nil {
		return ks, err
	}

	ks = PublicKeySet{
		Expiry: TimeNow().Add(ttl),
		Keys:   map[string]*rsa.PublicKey{},
	}

	for _, key := range keys.Keys {
		// we only plan on using RSA
		if key.Use == "sig" && key.Kty == "RSA" {
			n, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				return ks, err
			}
			e, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				return ks, err
			}
			ei := big.NewInt(0).SetBytes(e).Int64()
			ks.Keys[key.Kid] = &rsa.PublicKey{
				N: big.NewInt(0).SetBytes(n),
				E: int(ei),
			}
		}
	}
	return ks, nil
}

// TimeNow is used internally to determine the current time. It has been abstracted to
// this global function as a mechanism to help with testing.
var TimeNow = func() time.Time { return time.Now() }
