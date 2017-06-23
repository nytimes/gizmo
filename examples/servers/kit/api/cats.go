package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NYTimes/gizmo/server/kit"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	ocontext "golang.org/x/net/context"

	"github.com/NYTimes/gizmo/examples/nyt"
)

// GRPC layer, add the service-wide middleware ourselves
func (s service) GetCats(ctx ocontext.Context, r *google_protobuf.Empty) (*CatsResponse, error) {
	res, err := s.getCats(ctx, r)
	if res != nil {
		return res.(*CatsResponse), err
	}
	return nil, err
}

// SHARED BUSINESS LAYER
func (s service) getCats(ctx context.Context, _ interface{}) (interface{}, error) {
	res, err := s.client.SemanticConceptSearch("des", "cats")
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to get cats")
		return nil, kit.NewJSONStatusResponse(
			&CatsResponse{Status: "ERROR"},
			http.StatusInternalServerError)
	}
	kit.LogMsg(ctx, fmt.Sprintf("cats results found: %d", len(res)))
	return semToCat(res), nil
}

// BIZ LOGIC (SHOULD/COULD BE IN SOME BIZ PACKAGE)
func semToCat(res []*nyt.SemanticConceptArticle) *CatsResponse {
	var cs CatsResponse
	// translate Semantic to CatResponse
	if res != nil {
		cs.NumResults = uint32(len(res))
		cs.Status = "OK"
		cs.Results = make([]*CatResult, len(res))
		for i, a := range res {
			cs.Results[i] = &CatResult{
				Title:  a.Title,
				URL:    a.Url,
				Byline: a.Byline,
				Body:   a.Body,
			}
		}
	}
	return &cs
}
