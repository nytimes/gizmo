package appengine

import (
	"net/http"

	"github.com/NYTimes/gizmo/web"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func (s *AppEngineService) GetMostPopular(ctx context.Context, r *http.Request) (int, interface{}, error) {
	resourceType := mux.Vars(r)["resourceType"]
	section := mux.Vars(r)["section"]
	timeframe := web.GetUInt64Var(r, "timeframe")
	res, err := nytclient().GetMostPopular(ctx, resourceType, section, uint(timeframe))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, res, nil
}
