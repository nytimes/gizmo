package readinglist

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	ocontext "golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/nytimes/gizmo/server/kit"
)

// gRPC stub
func (s service) PutLink(ctx ocontext.Context, r *PutLinkRequest) (*Message, error) {
	res, err := s.putLink(ctx, r)
	if err != nil {
		return nil, err
	}
	return res.(*Message), nil
}

// go-kit endpoint.Endpoint with core business logic
func (s service) putLink(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*PutLinkRequest)

	// validate the request
	if !strings.HasPrefix(r.Request.Link.Url, "https://www.nytimes.com/") {
		return nil, kit.NewJSONStatusResponse(
			&Message{"only https://www.nytimes.com URLs accepted"},
			http.StatusBadRequest)
	}

	var err error
	// call the service-injected DB interface
	if r.Request.Delete {
		err = s.db.DeleteLink(ctx, getUser(ctx), r.Request.Link.Url)
	} else {
		err = s.db.PutLink(ctx, getUser(ctx), r.Request.Link.Url)
	}
	if err != nil {
		return nil, kit.NewJSONStatusResponse(
			&Message{"problems updating link"},
			http.StatusInternalServerError)
	}

	return &Message{Message: "success"}, nil
}

// JSON request decoder
func decodePutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var lr LinkRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil || lr.Link == nil {
		return nil, kit.NewJSONStatusResponse(
			&Message{Message: "bad request"},
			http.StatusBadRequest)
	}
	return &PutLinkRequest{
		Request: &lr}, nil
}

// Protobuf request decoder
func decodePutProtoRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, kit.NewJSONStatusResponse(
			&Message{Message: "unable to read request"},
			http.StatusBadRequest)
	}
	var lr LinkRequest
	err = proto.Unmarshal(b, &lr)
	if err != nil {
		return nil, kit.NewJSONStatusResponse(
			&Message{Message: "bad request"},
			http.StatusBadRequest)
	}
	return &PutLinkRequest{
		Request: &lr}, nil
}
