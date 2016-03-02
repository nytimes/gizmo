package service

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// Delete is JSONEndpoint for deleting a saved item from a user's list.
func (s *SavedItemsService) Delete(ctx context.Context, r *http.Request) (int, interface{}, error) {
	// gather the inputs from request
	var usr *user.User
	if usr = user.Current(ctx); usr == nil {
		return http.StatusUnauthorized, nil, errors.New("please visit /svc/login before accessing saved items")
	}
	url := r.URL.Query().Get("url")

	// do work and respond
	err := s.repo.Delete(ctx, usr.ID, url)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	log.Infof(ctx, "successfully deleted item")
	return http.StatusOK, jsonResponse{"successfully deleted saved item"}, nil
}
