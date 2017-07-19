package readinglist

import (
	"context"
	"net/http"
	"strconv"

	ocontext "golang.org/x/net/context"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/pkg/errors"
)

// gRPC stub
func (s service) GetListLimit(ctx ocontext.Context, r *GetListLimitRequest) (*Links, error) {
	res, err := s.getLinks(ctx, r)
	if err != nil {
		return nil, err
	}
	return res.(*Links), nil
}

// go-kit endpoint.Endpoint with core business logic
func (s service) getLinks(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*GetListLimitRequest)

	// set request defaults
	if r.Limit == 0 {
		r.Limit = 50
	}

	// get data from the service-injected DB interface
	links, err := s.db.GetLinks(ctx, getUser(ctx), int(r.Limit))
	if err != nil {
		kit.LogErrorMsgWithFields(ctx, err, "error getting links from DB")
		return nil, kit.NewJSONStatusResponse(
			&Message{"server error"},
			http.StatusInternalServerError)
	}
	lks := make([]*Link, len(links))
	for i, l := range links {
		lks[i] = &Link{Url: l}
	}
	return &Links{Links: lks}, errors.Wrap(err, "unable to get links")
}

// request decoder can be used for proto and JSON since there is no body
func decodeGetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	limit, err := strconv.ParseInt(kit.Vars(r)["limit"], 10, 64)
	if err != nil {
		return nil, kit.NewJSONStatusResponse(
			&Message{"bad request"},
			http.StatusBadRequest)
	}
	return &GetListLimitRequest{
		Limit: int32(limit),
	}, nil
}
