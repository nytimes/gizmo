package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/oauth2/jws"
)

type Verifier struct {
	ks PublicKeySource
	df ClaimsDecoderFunc
	vf VerifyFunc

	skewAllowance int64
}

var defaultSkewAllowance = time.Minute * 5

type ClaimSetter interface {
	BaseClaims() *jws.ClaimSet
}

type ClaimsDecoderFunc func(context.Context, []byte) (ClaimSetter, error)

type VerifyFunc func(context.Context, interface{}) bool

func NewVerifier(ks PublicKeySource, df ClaimsDecoderFunc, vf VerifyFunc) *Verifier {
	return &Verifier{
		ks:            ks,
		df:            df,
		vf:            vf,
		skewAllowance: int64(defaultSkewAllowance.Seconds()),
	}
}

func (c Verifier) Verify(ctx context.Context, token string) (bool, error) {
	// decode token to get header
	hdr, rawPayload, err := decodeToken(token)
	if err != nil {
		return false, err
	}

	// get keyset
	keys, err := c.ks.Get(ctx)
	if err != nil {
		return false, err
	}

	// get key from keyset
	key, err := keys.GetKey(hdr.KeyID)
	if err != nil {
		return false, err
	}

	// verify token
	err = jws.Verify(token, key)
	if err != nil {
		return false, err
	}

	// use claims decoder func
	clmstr, err := c.df(ctx, rawPayload)
	if err != nil {
		return false, err
	}

	claims := clmstr.BaseClaims()
	nowUnix := timeNow().Unix()

	if nowUnix < (claims.Iat - c.skewAllowance) {
		return false, errors.New("invalid issue time")
	}

	if nowUnix > (claims.Exp + c.skewAllowance) {
		return false, errors.New("invalid expiration time")
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
