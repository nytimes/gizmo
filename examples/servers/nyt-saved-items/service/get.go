package service

import (
	"net/http"

	"github.com/gorilla/context"
)

func (s *SavedItemsService) Get(r *http.Request) (int, interface{}, error) {
	// gather the input
	id := context.Get(r, userIDKey).(uint64)

	// do work and respond
	items, err := s.repo.Get(id)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, items, nil
}
