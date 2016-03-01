package appengine

import (
	"net/http"

	"golang.org/x/net/context"
)

func (s *AppEngineService) GetCats(ctx context.Context, r *http.Request) (int, interface{}, error) {
	res, err := s.client.SemanticConceptSearch(ctx, "des", "cats")
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, res, nil
}
