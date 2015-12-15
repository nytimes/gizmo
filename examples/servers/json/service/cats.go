package service

import "net/http"

func (s *JSONService) GetCats(r *http.Request) (int, interface{}, error) {
	res, err := s.client.SemanticConceptSearch("des", "cats")
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, res, nil
}
