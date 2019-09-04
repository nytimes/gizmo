package kit

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// EncodeProtoRequest is an httptransport.EncodeRequestFunc that serializes the request
// as Protobuf. If the request implements Headerer, the provided headers will be applied
// to the request.
func EncodeProtoRequest(_ context.Context, r *http.Request, preq interface{}) error {
	r.Header.Set("Content-Type", "application/octet-stream")
	if headerer, ok := preq.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	req, ok := preq.(proto.Message)
	if !ok {
		return errors.New("response does not implement proto.Message")
	}

	b, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	r.ContentLength = int64(len(b))
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	return nil
}
