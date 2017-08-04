// +build go1.7

package service

import (
	"context"
	"net/http"

	ocontext "golang.org/x/net/context"

	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/server"
)

func (s *RPCService) GetCats(ctx ocontext.Context, r *CatsRequest) (*CatsResponse, error) {
	var (
		err error
		res []*nyt.SemanticConceptArticle
	)
	defer server.MonitorRPCRequest()(ctx, "GetCats", &err)

	res, err = s.client.SemanticConceptSearch("des", "cats")
	if err != nil {
		return nil, err
	}

	return &CatsResponse{res}, nil
}

func (s *RPCService) GetCatsJSON(ctx context.Context, r *http.Request) (int, interface{}, error) {
	res, err := s.GetCats(ctx, &CatsRequest{})
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, res.Results, nil
}
