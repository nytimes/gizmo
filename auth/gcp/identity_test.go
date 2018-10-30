package gcp

import (
	"testing"
)

func TestIDKeySource(t *testing.T) {
	/*
		ctx := context.Background()
		ks, err := NewIdentityPublicKeySource(ctx)
		if err != nil {
			t.Errorf("unable to init key source: %s", err)
			return
		}

		ts, err := NewIdentityTokenSource("nytimes.com")
		if err != nil {
			t.Errorf("unable to init token source: %s", err)
			return
		}

		tkn, err := ts.Token()
		if err != nil {
			t.Errorf("unable to get token: %s", err)
			return
		}

		v := auth.NewVerifier(ks, IdentityClaimsDecoderFunc,
			func(ctx context.Context, c interface{}) bool {
				ic, ok := c.(IdentityClaimSet)
				if !ok {
					t.Errorf("invalid claims type: %T", c)
					return false
				}

				t.Logf("claims: %#v", ic)
				return true
			})

		good, err := v.Verify(ctx, tkn.AccessToken)
		if err != nil {
			t.Errorf("unable to verify token: %s", err)
			return
		}

		if !good {
			t.Error("token was not verified")
		}
	*/
}
