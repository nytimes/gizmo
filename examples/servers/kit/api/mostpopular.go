package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/NYTimes/gizmo/server/kit"
	ocontext "golang.org/x/net/context"

	"github.com/NYTimes/gizmo/examples/nyt"
)

// GRPC LAYER, add the middleware layer ourselves
func (s service) GetMostPopularResourceTypeSectionTimeframe(ctx ocontext.Context, req *GetMostPopularResourceTypeSectionTimeframeRequest) (*MostPopularResponse, error) {
	res, err := s.Middleware(s.getMostPopular)(ctx, req)
	if res != nil {
		return res.(*MostPopularResponse), err
	}
	return nil, err
}

// SHARED BIZ LAYER
func (s service) getMostPopular(ctx context.Context, r interface{}) (interface{}, error) {
	mpr := r.(*GetMostPopularResourceTypeSectionTimeframeRequest)

	res, err := s.client.GetMostPopular(mpr.ResourceType, mpr.Section, uint(mpr.Timeframe))
	if err != nil {
		return nil, kit.NewJSONStatusResponse(
			&GetMostPopularResourceTypeSectionTimeframeRequest{},
			http.StatusBadRequest)
	}

	kit.LogMsg(ctx, fmt.Sprintf("most popular results found: %d", len(res)))
	return mpToMP(res), nil
}

// CUSTOM HTTP REQUEST DECODER
func decodeMostPopularRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vs := kit.Vars(r)
	timeframe, err := strconv.ParseUint(vs["timeframe"], 10, 8)
	if err != nil {
		return nil, kit.NewJSONStatusResponse(
			&MostPopularResponse{Status: "bad request"},
			http.StatusBadRequest)
	}
	return &GetMostPopularResourceTypeSectionTimeframeRequest{
		ResourceType: vs["resourceType"],
		Section:      vs["section"],
		Timeframe:    int32(timeframe),
	}, nil
}

// BIZ LOGIC THAT SHOULD/COULD LIVE SOMEWHERE ELSE?
func mpToMP(res []*nyt.MostPopularResult) *MostPopularResponse {
	var mpr MostPopularResponse
	mpr.NumResults = uint32(len(res))
	mpr.Status = "OK"
	mpr.Results = make([]*MostPopularResult, len(res))
	for i, r := range res {
		mpr.Results[i] = &MostPopularResult{
			Abstract:      r.Abstract,
			AssetID:       r.AsssetId,
			Byline:        r.Byline,
			Column:        r.Column,
			ID:            r.Id,
			Keywords:      r.AdxKeywords,
			PublishedDate: r.PublishedDate,
			Section:       r.Section,
			Source:        r.Source,
			Title:         r.Title,
			Type:          r.Type,
			URL:           r.Url,
		}
	}
	return &mpr
}
