package service

import (
	"net/http"

	"go.pedge.io/google-protobuf"

	"github.com/nytimes/gizmo/examples/nyt"
	"github.com/nytimes/gizmo/server"
	"golang.org/x/net/context"
)

func (s *RPCService) GetCats(ctx context.Context, r *google_protobuf.Empty) (*CatsResponse, error) {
	var (
		err error
		res []*nyt.SemanticConceptArticle
	)
	defer server.MonitorRPCRequest()(ctx, "GetCats", err)

	res, err = s.client.SemanticConceptSearch("des", "cats")
	if err != nil {
		return nil, err
	}

	return &CatsResponse{res}, nil
}

func (s *RPCService) GetCatsJSON(r *http.Request) (int, interface{}, error) {
	res, err := s.GetCats(context.Background(), google_protobuf.EmptyInstance)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, res.Result, nil
}
