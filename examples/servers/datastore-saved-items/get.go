package service

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/user"
)

// Get is a JSONEndpoint to return a list of saved items for the given user ID.
func (s *SavedItemsService) Get(ctx context.Context, r *http.Request) (int, interface{}, error) {
	// gather the input from the request
	var usr *user.User
	if usr = user.Current(ctx); usr == nil {
		return http.StatusUnauthorized, nil, errors.New("please visit /svc/login before accessing saved items")
	}

	// do work and respond
	items, err := s.repo.Get(ctx, usr.ID)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, items, nil
}
