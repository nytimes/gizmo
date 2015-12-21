package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/web"
	"github.com/gorilla/mux"
)

func (s *SimpleService) GetMostPopular(r *http.Request) (int, interface{}, error) {
	resourceType := mux.Vars(r)["resourceType"]
	section := mux.Vars(r)["section"]
	timeframe := web.GetUInt64Var(r, "timeframe")
	res, err := s.client.GetMostPopular(resourceType, section, uint(timeframe))
	if err != nil {
		return http.StatusInternalServerError, nil, &jsonErr{err.Error()}
	}
	return http.StatusOK, res, nil
}

type jsonErr struct {
	Err string `json:"error"`
}

func (e *jsonErr) Error() string {
	return e.Err
}
