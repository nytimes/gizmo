package auth

import (
	"context"
	"crypto/rsa"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestResuseKeySource(t *testing.T) {
	testTime := time.Date(2018, 10, 29, 12, 0, 0, 0, time.UTC)

	TimeNow = func() time.Time { return testTime }

	firstKeys, err := NewPublicKeySetFromJSON([]byte(testGoogleCerts), 1*time.Second)
	if err != nil {
		t.Errorf("unexpected error creating key set: %s", err)
		return
	}

	nextKeys := PublicKeySet{
		Expiry: testTime.Add(2 * time.Second),
		Keys: map[string]*rsa.PublicKey{
			"8289d54280b76712de41cd2ef95972b123be9ac0": &rsa.PublicKey{N: testGoogle(testGoogleKey2), E: 65537},
		},
	}

	reuser := NewReusePublicKeySource(firstKeys, testKeySource{keys: nextKeys})

	// first get, firstKeys are not expired and should be returned

	gotKeys, err := reuser.Get(context.Background())
	if err != nil {
		t.Errorf("unexpected error getting keys: %s", err)
		return
	}
	if !cmp.Equal(gotKeys, firstKeys, cmpopts.IgnoreUnexported(big.Int{})) {
		t.Errorf("first keys did not match expectations: %s", cmp.Diff(gotKeys, firstKeys,
			cmpopts.IgnoreUnexported(big.Int{})))
		return
	}

	// move time forward, expire the first keys
	TimeNow = func() time.Time { return testTime.Add(1500 * time.Millisecond) }

	gotKeys, err = reuser.Get(context.Background())
	if err != nil {
		t.Errorf("unexpected error getting keys: %s", err)
		return
	}
	if !cmp.Equal(gotKeys, nextKeys, cmpopts.IgnoreUnexported(big.Int{})) {
		t.Errorf("next keys did not match expectations: %s", cmp.Diff(gotKeys, nextKeys,
			cmpopts.IgnoreUnexported(big.Int{})))
		return
	}

	// verify get works
	k, err := gotKeys.GetKey("8289d54280b76712de41cd2ef95972b123be9ac0")
	if err != nil {
		t.Errorf("unexpected error getting key: %s", err)
	}
	if !cmp.Equal(k, nextKeys.Keys["8289d54280b76712de41cd2ef95972b123be9ac0"],
		cmpopts.IgnoreUnexported(big.Int{})) {
		t.Errorf("next keys did not match expectations: %s",
			cmp.Diff(k, nextKeys.Keys["8289d54280b76712de41cd2ef95972b123be9ac0"],
				cmpopts.IgnoreUnexported(big.Int{})))
		return
	}
}

type testKeySource struct {
	keys PublicKeySet
	err  error
}

func (t testKeySource) Get(ctx context.Context) (PublicKeySet, error) {
	return t.keys, t.err
}

func TestKeySetFromURL(t *testing.T) {
	testTime := time.Date(2018, 10, 29, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string

		givenCacheHeader string
		givenPayload     string
		givenBadServer   bool

		wantError bool
		wantKeys  PublicKeySet
	}{
		{
			name: "normal google certs w/ cache",

			givenCacheHeader: "max-age=1",
			givenPayload:     testGoogleCerts,

			wantKeys: PublicKeySet{
				Expiry: testTime.Add(1 * time.Second),
				Keys: map[string]*rsa.PublicKey{
					"728f4016652079b9ed99861bb09bafc5a45baa86": {N: testGoogle(testGoogleKey1), E: 65537},
					"8289d54280b76712de41cd2ef95972b123be9ac0": {N: testGoogle(testGoogleKey2), E: 65537},
				},
			},
		},
		{
			name: "normal google certs w/o cache",

			givenPayload: testGoogleCerts,

			wantKeys: PublicKeySet{
				Expiry: testTime.Add(5 * time.Second),
				Keys: map[string]*rsa.PublicKey{
					"728f4016652079b9ed99861bb09bafc5a45baa86": {N: testGoogle(testGoogleKey1), E: 65537},
					"8289d54280b76712de41cd2ef95972b123be9ac0": {N: testGoogle(testGoogleKey2), E: 65537},
				},
			},
		},
		{
			name:         "bad response, want error",
			givenPayload: "<html>some kind of angry robot</html>",

			wantError: true,
		},
		{
			name: "bad server, want error",

			givenPayload:   testGoogleCerts,
			givenBadServer: true,

			wantError: true,
		},
		{
			name: "bad cache header, want success",

			givenCacheHeader: "BLAARB! ANGRY COMPUTER!",
			givenPayload:     testGoogleCerts,

			wantKeys: PublicKeySet{
				Expiry: testTime.Add(5 * time.Second),
				Keys: map[string]*rsa.PublicKey{
					"728f4016652079b9ed99861bb09bafc5a45baa86": {N: testGoogle(testGoogleKey1), E: 65537},
					"8289d54280b76712de41cd2ef95972b123be9ac0": {N: testGoogle(testGoogleKey2), E: 65537},
				},
			},
		},
	}

	for _, test := range tests {
		TimeNow = func() time.Time { return testTime }

		t.Run(test.name, func(t *testing.T) {
			srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.givenCacheHeader != "" {
					w.Header().Add("cache-control", test.givenCacheHeader)
				}
				w.Write([]byte(test.givenPayload))
			}))
			defer srvr.Close()
			if test.givenBadServer {
				srvr.Close()
			}

			got, gotErr := NewPublicKeySetFromURL(http.DefaultClient, srvr.URL, 5*time.Second)
			if (gotErr != nil) != test.wantError {
				t.Errorf("expected error? %t and got %s", test.wantError, gotErr)
				return
			}

			if !cmp.Equal(got, test.wantKeys, cmpopts.IgnoreUnexported(big.Int{})) {
				t.Errorf("keys did not match expectations: %s", cmp.Diff(got, test.wantKeys, cmpopts.IgnoreUnexported(big.Int{})))
			}
		})
	}
}

var (
	testGoogleKey1 = "22433090823316839640339489484457787676134304275873755218133861343920545237994470293495919014803004482856016084553850209153845425382613518932089311310596313310600424737736088033780907099977873221447195709312051528384355479077579673777886481089832045696620374920724411025483234264634539593436130076854768802102666090698524255278976644754677212286402099970599598264338136458077064875129043902522602870213617296706363155049264877048351659848686562003749244021217935734825983116131356048732262346697829165992404416525006735905763408678841171079087251498194471555953631250995421460080193870950448459655537601409828979901677"
	testGoogleKey2 = "18112684417237113466774220553948287658642275536612278117358654328223325254239855900914880208928002193971790755769647369251185307648390933383803629253833792935549104394492595490970480288704258432536877269087694080352968836583401030682357884420432445619092471675752640354212779048186101852385524325549753366939320751885000360016238872619721767196169731422128756698973826778639560486979276112061913581353475855995717107174242233057925781337843224898645582603363390951368105740797845693907662079988391116580563176804122832211438500322243675724500523751141979116987975024595515232643410130766424608026731615327022863391377"
)

func testGoogle(testKey string) *big.Int {
	i, ok := new(big.Int).SetString(testKey, 10)
	if !ok {
		panic("bad number: " + testKey)
	}
	return i
}

const testGoogleCerts = `{
  "keys": [
    {
      "alg": "RS256",
      "n": "j3rnumgAbXUMwqwL31lNVYuvWGbriT7uy7CgTeqwfNLf6Q9TMQlDidrFFSLgSe2BifjEswS6B4qpsXMlrMoSozIwbHuSkoQZdY2m5vFEZRkyHB4mAKZuzUi5XH5LVllbv2TBp6KsjJWSSW5Bnyen9pIeamHdODoX_PEdBhmqknDURRuuq3Bb3IVnudGP8JCbHTZ86ZS2aS2hpYK3eA8dvp155K1jKMAG9WH89jhkeFR67Oq9mD-yGZDaNCN8nOZR5Iyw-WQJo5-ijEAckHXn1SdYGjQgm_2fvEBsf0gJEjmx2DrNeacZLyDUA_dB9JIy5ZrfFZ0H9l0IkSgggUSakQ",
      "use": "sig",
      "kid": "8289d54280b76712de41cd2ef95972b123be9ac0",
      "e": "AQAB",
      "kty": "RSA"
    },
    {
      "use": "sig",
      "kid": "728f4016652079b9ed99861bb09bafc5a45baa86",
      "e": "AQAB",
      "kty": "RSA",
      "alg": "RS256",
      "n": "sbRNoYaZX3w2Iosb6uzfykt-uJh_NRVQ0h_98Gptkpq3r-xgdaq9i-mmZEYZtrNUmIqOEDvtIJ36-CVnDZI2p_eARFkmedHC14QX5SHdFb2qr0a5DuqC5qLoyOMXSNJyfRHK8ULjozLxO7t_P0EsdlLPOUQjcbpTiIo9p-L9iskMCKpQdDfQ4CrzHKQjfYN3KJdehsChguffue-VBUkoDaRRUA50h6DiFe-loC_dzycoNGYJEJvAM5DC3zuHr6dfc5saHLUi4upgR2_jchA6kwSOVBC05qUgY4E3UdYTWciTqkSowiAErDx21g-oB6QzIr8MRMzKa89-g2Ine-qE7Q"
    }
  ]
}`
