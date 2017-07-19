package kit

import (
	"encoding/json"
	"net/http"
)

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
