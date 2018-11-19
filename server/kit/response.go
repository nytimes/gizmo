package kit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/golang/protobuf/proto"
)

// NewProtoStatusResponse allows users to respond with a specific HTTP status code and
// a Protobuf or JSON serialized response.
func NewProtoStatusResponse(res proto.Message, code int) *ProtoStatusResponse {
	return &ProtoStatusResponse{res: res, code: code}
}

// ProtoStatusResponse implements:
// `httptransport.StatusCoder` to allow users to respond with the given
// response with a non-200 status code.
// `proto.Marshaler` and proto.Message so it can wrap a proto Endpoint responses.
// `json.Marshaler` so it can wrap JSON Endpoint responses.
// `error` so it can be used to respond as an error within the go-kit stack.
type ProtoStatusResponse struct {
	code int
	res  proto.Message
}

// StatusCode implements httptransport.StatusCoder and will return the given HTTP code.
func (c *ProtoStatusResponse) StatusCode() int {
	return c.code
}

// Marshal is to implement proto.Marshaler. It will marshal the given message, not this
// struct.
func (c *ProtoStatusResponse) Marshal() ([]byte, error) {
	return proto.Marshal(c.res)
}

// Reset is to implement proto.Message. It uses the given message's Reset method.
func (c *ProtoStatusResponse) Reset() { c.res.Reset() }

// String is to implement proto.Message. It uses the given message's String method.
func (c *ProtoStatusResponse) String() string { return c.res.String() }

// ProtoMessage is to implement proto.Message. It uses the given message's ProtoMessage
// method.
func (c *ProtoStatusResponse) ProtoMessage() { c.res.ProtoMessage() }

// MarshalJSON is to implement json.Marshaler. It will marshal the given message, not
// this struct.
func (c *ProtoStatusResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.res)
}

var _ proto.Marshaler = &ProtoStatusResponse{}

// to implement error
func (c *ProtoStatusResponse) Error() string {
	return http.StatusText(c.code)
}

// NewJSONStatusResponse allows users to respond with a specific HTTP status code and
// a JSON serialized response.
func NewJSONStatusResponse(res interface{}, code int) *JSONStatusResponse {
	return &JSONStatusResponse{res: res, code: code}
}

// JSONStatusResponse implements:
// `httptransport.StatusCoder` to allow users to respond with the given
// response with a non-200 status code.
// `json.Marshaler` so it can wrap JSON Endpoint responses.
// `error` so it can be used to respond as an error within the go-kit stack.
type JSONStatusResponse struct {
	code int
	res  interface{}
}

// StatusCode is to implement httptransport.StatusCoder
func (c *JSONStatusResponse) StatusCode() int {
	return c.code
}

// MarshalJSON is to implement json.Marshaler
func (c *JSONStatusResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.res)
}

// Error is to implement error
func (c *JSONStatusResponse) Error() string {
	return http.StatusText(c.code)
}

// EncodeProtoResponse is an httptransport.EncodeResponseFunc that serializes the response
// as Protobuf. Many Proto-over-HTTP services can use it as a sensible default. If the
// response implements Headerer, the provided headers will be applied to the response.
// If the response implements StatusCoder, the provided StatusCode will be used instead
// of 200.
func EncodeProtoResponse(ctx context.Context, w http.ResponseWriter, pres interface{}) error {
	res, ok := pres.(proto.Message)
	if !ok {
		return errors.New("response does not implement proto.Message")
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	if headerer, ok := w.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	code := http.StatusOK
	if sc, ok := pres.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}
	if res == nil {
		return nil
	}
	b, err := proto.Marshal(res)
	if err != nil {
		// maybe log instead? need to avoid a second header write
		return nil
	}
	_, err = w.Write(b)
	if err != nil {
		// maybe log instead? need to avoid a second header write
	}
	return nil
}
