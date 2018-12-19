package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
)

func (s *MixedService) GetMostPopular(r *http.Request) (int, interface{}, error) {
	resourceType := server.Vars(r)["resourceType"]
	section := server.Vars(r)["section"]
	timeframe := server.GetUInt64Var(r, "timeframe")
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
