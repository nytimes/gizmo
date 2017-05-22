package httpsvc

import (
	"context"
	"net/http"
	"strconv"

	"github.com/NYTimes/gizmo/server/kithttp"
)

func (s httpService) GetMostPopular(ctx context.Context, r interface{}) (interface{}, error) {
	mpr := r.(MostPopularRequest)

	res, err := s.client.GetMostPopular(mpr.ResourceType, mpr.Section, mpr.Timeframe)
	if err != nil {
		return nil, kithttp.NewJSONStatusResponse(err.Error(), http.StatusBadRequest)
	}
	kithttp.Logger(ctx).Log("most popular results found", len(res))
	return res, nil
}

func decodeMostPopularRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vs := kithttp.Vars(r)
	timeframe, err := strconv.ParseUint(vs["timeframe"], 10, 8)
	if err != nil {
		return nil, kithttp.NewJSONStatusResponse(
			"unable to parse timeframe value: "+err.Error(),
			http.StatusBadRequest)
	}

	return MostPopularRequest{
		ResourceType: vs["resourceType"],
		Section:      vs["section"],
		Timeframe:    uint(timeframe),
	}, nil
}

type MostPopularRequest struct {
	ResourceType string
	Section      string
	Timeframe    uint
}
