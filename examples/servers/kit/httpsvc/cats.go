package httpsvc

import (
	"context"
	"net/http"

	"github.com/NYTimes/gizmo/server/kithttp"
)

func (s httpService) GetCats(ctx context.Context, _ interface{}) (interface{}, error) {
	res, err := s.client.SemanticConceptSearch("des", "cats")
	if err != nil {
		return nil, kithttp.NewJSONStatusResponse(err.Error(),
			http.StatusInternalServerError)
	}
	kithttp.Logger(ctx).Log("cats results found", len(res))
	return res, nil
}
