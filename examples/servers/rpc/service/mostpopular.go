package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/examples/nyt"
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gizmo/web"
	"golang.org/x/net/context"
)

func (s *RPCService) GetMostPopular(ctx context.Context, r *MostPopularRequest) (*MostPopularResponse, error) {
	var (
		err error
		res []*nyt.MostPopularResult
	)
	defer server.MonitorRPCRequest()(ctx, "GetMostPopular", err)

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
