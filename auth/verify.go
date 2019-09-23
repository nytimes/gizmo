package auth // import "github.com/NYTimes/gizmo/auth"

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/jws"
)

// Verifier is a generic tool for verifying JWT tokens.
type Verifier struct {
	ks PublicKeySource
	df ClaimsDecoderFunc
	vf VerifyFunc

	skewAllowance int64
}

// ErrBadCreds will always be wrapped when a user's
// credentials are unexpected. This is so that we can
// distinguish between a client error from a server error
var ErrBadCreds = errors.New("bad credentials")

var defaultSkewAllowance = time.Minute * 5

// ClaimSetter is an interface for all incoming claims to implement. This ensures the
// basic format used by the `jws` package.
type ClaimSetter interface {
	BaseClaims() *jws.ClaimSet
}

// ClaimsDecoderFunc will expect to convert a JSON payload into the appropriate claims
// type.
type ClaimsDecoderFunc func(context.Context, []byte) (ClaimSetter, error)

// VerifyFunc will be called by the Verify if all other checks on the token pass.
// Developers should use this to encapsulate any business logic involved with token
// verification.
type VerifyFunc func(context.Context, interface{}) bool

// NewVerifier returns a genric Verifier that will use the given funcs and key source.
func NewVerifier(ks PublicKeySource, df ClaimsDecoderFunc, vf VerifyFunc) *Verifier {
	return &Verifier{
		ks:            ks,
		df:            df,
		vf:            vf,
		skewAllowance: int64(defaultSkewAllowance.Seconds()),
	}
}

// VerifyInboundKitContext is meant to be used within a go-kit stack that has populated
// the context with common headers, specficially
// kit/transport/http.ContextKeyRequestAuthorization.
func (c Verifier) VerifyInboundKitContext(ctx context.Context) (bool, error) {
	authHdr, ok := ctx.Value(httptransport.ContextKeyRequestAuthorization).(string)
	if !ok {
		return false, errors.New("auth header did not exist")
	}

	token, err := parseHeader(authHdr)
	if err != nil {
		return false, err
	}

	return c.Verify(ctx, token)
}

// VerifyRequest will pull the token from the "Authorization" header of the inbound
// request then decode and verify it.
func (c Verifier) VerifyRequest(r *http.Request) (bool, error) {
	token, err := GetAuthorizationToken(r)
	if err != nil {
		return false, err
	}

	return c.Verify(r.Context(), token)
}

// Verify will accept an opaque JWT token, decode it and verify it.
func (c Verifier) Verify(ctx context.Context, token string) (bool, error) {
	hdr, rawPayload, err := decodeToken(token)
	if err != nil {
		return false, errors.Wrap(ErrBadCreds, err.Error())
	}

	keys, err := c.ks.Get(ctx)
	if err != nil {
		return false, err
	}

	key, err := keys.GetKey(hdr.KeyID)
	if err != nil {
		return false, err
	}

	err = jws.Verify(token, key)
	if err != nil {
		return false, errors.Wrap(ErrBadCreds, err.Error())
	}

	// use claims decoder func
	clmstr, err := c.df(ctx, rawPayload)
	if err != nil {
		return false, err
	}

	claims := clmstr.BaseClaims()
	nowUnix := TimeNow().Unix()

	if nowUnix < (claims.Iat - c.skewAllowance) {
		return false, errors.New("invalid issue time")
	}

	if nowUnix > (claims.Exp + c.skewAllowance) {
		return false, errors.Wrap(ErrBadCreds, "invalid expiration time")
	}

	return c.vf(ctx, clmstr), nil
}

func decodeToken(token string) (*jws.Header, []byte, error) {
	s := strings.Split(token, ".")
	if len(s) != 3 {
		return nil, nil, errors.New("invalid token")
	}

	dh, err := base64.RawURLEncoding.DecodeString(s[0])
	if err != nil {
		return nil, nil, err
	}
	var h jws.Header
	err = json.Unmarshal(dh, &h)
	if err != nil {
		return nil, nil, err
	}

	dcs, err := base64.RawURLEncoding.DecodeString(s[1])
	if err != nil {
		return nil, nil, err
	}
	return &h, dcs, nil
}

func parseHeader(hdr string) (string, error) {
	auths := strings.Split(hdr, " ")
	if len(auths) != 2 {
		return "", errors.New("auth header invalid format")
	}
	return auths[1], nil
}

// GetAuthorizationToken will pull the Authorization header from the given request and
// attempt to retrieve the token within it.
func GetAuthorizationToken(r *http.Request) (string, error) {
	return parseHeader(r.Header.Get("Authorization"))
}
