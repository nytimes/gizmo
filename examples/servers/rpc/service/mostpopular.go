// +build go1.7

package service

import (
	"context"
	"net/http"

	ocontext "golang.org/x/net/context"

	"github.com/nytimes/gizmo/examples/nyt"
	"github.com/nytimes/gizmo/server"
	"github.com/nytimes/gizmo/web"
)

func (s *RPCService) GetMostPopular(ctx ocontext.Context, r *MostPopularRequest) (*MostPopularResponse, error) {
	var (
		err error
		res []*nyt.MostPopularResult
	)
	defer server.MonitorRPCRequest()(ctx, "GetMostPopular", &err)

	res, err = s.client.GetMostPopular(r.ResourceType, r.Section, uint(r.TimePeriodDays))
	if err != nil {
		return nil, err
	}
	return &MostPopularResponse{res}, nil
}

func (s *RPCService) GetMostPopularJSON(ctx context.Context, r *http.Request) (int, interface{}, error) {
	res, err := s.GetMostPopular(
		ctx,
		&MostPopularRequest{
			web.Vars(r)["resourceType"],
			web.Vars(r)["section"],
			uint32(web.GetUInt64Var(r, "timeframe")),
		})
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, res.Result, nil
}
