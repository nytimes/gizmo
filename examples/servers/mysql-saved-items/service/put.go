package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/gorilla/context"
)

// Put is a JSONEndpoint for adding a new saved item to a user's list.
func (s *SavedItemsService) Put(r *http.Request) (int, interface{}, error) {
	// gather the inputs from the request
	id := context.Get(r, userIDKey).(uint64)
	url := r.URL.Query().Get("url")

	// do work and respond
	err := s.repo.Put(id, url)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	server.LogWithFields(r).Info("successfully saved item")
	return http.StatusCreated, jsonResponse{"successfully saved item"}, nil
}
