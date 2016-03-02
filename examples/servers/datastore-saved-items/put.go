package service

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// Put is a JSONEndpoint for adding a new saved item to a user's list.
func (s *SavedItemsService) Put(ctx context.Context, r *http.Request) (int, interface{}, error) {
	// gather the inputs from the request
	var usr *user.User
	if usr = user.Current(ctx); usr == nil {
		return http.StatusUnauthorized, nil, errors.New("please visit /svc/login before accessing saved items")
	}
	url := r.FormValue("url")
	// do work and respond
	err := s.repo.Put(ctx, usr.ID, url)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	log.Infof(ctx, "successfully saved item")
	return http.StatusCreated, jsonResponse{"successfully saved item"}, nil
}
