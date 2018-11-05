package gcp

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// GetDefaultEmail is a helper method for users on GCE or the 2nd generation GAE
// environment.
func GetDefaultEmail(ctx context.Context, addr string, hc *http.Client) (string, error) {
	email, err := metadataGet(ctx, addr, hc, "instance/service-accounts/default/email")
	return email, errors.Wrap(err, "unable to get default email from metadata")
}

func metadataGet(ctx context.Context, addr string, hc *http.Client, suffix string) (string, error) {
	if addr == "" {
		addr = "http://metadata/computeMetadata/v1/"
	}
	req, err := http.NewRequest(http.MethodGet, addr+suffix, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to create metadata request")
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := hc.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "unable to send request to metadata")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("metadata service returned a non-200 response: %d",
			resp.StatusCode)
	}

	tkn, err := ioutil.ReadAll(resp.Body)
	return string(tkn), errors.Wrap(err, "unable to read metadata response")
}
