package service

import (
	"net/http"

	"github.com/gorilla/context"
)

// Get is a JSONEndpoint to return a list of saved items for the given user ID.
func (s *SavedItemsService) Get(r *http.Request) (int, interface{}, error) {
	// gather the input from the request
	id := context.Get(r, userIDKey).(uint64)

	// do work and respond
	items, err := s.repo.Get(id)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, items, nil
}
