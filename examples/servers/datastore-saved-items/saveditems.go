package service

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/datastore"
)

type (
	// SavedItemsRepo is an interface layer between
	// our service and our database. Abstracting these methods
	// out of a pure implementation helps with testing.
	SavedItemsRepo interface {
		Get(context.Context, string) ([]*SavedItem, error)
		Put(context.Context, string, string) error
		Delete(context.Context, string, string) error
	}

	// DatastoreSavedItemsRepo is an implementation of the repo
	// interface built on top of Datastore.
	DatastoreSavedItemsRepo struct {
		kind    string
		dclient *datastore.Client
	}

	// SavedItem represents an article, blog, interactive, etc.
	// that a user wants to save for reading later.
	SavedItem struct {
		UserID    string    `json:"user_id"`
		URL       string    `json:"url"`
		Timestamp time.Time `json:"timestamp"`
	}
)

// NewSavedItemsRepo will attempt to connect to to Datastore and
// return a SavedItemsRepo implementation.
func NewSavedItemsRepo() SavedItemsRepo {
	return &DatastoreSavedItemsRepo{kind: "SavedItem"}
}

func (r *DatastoreSavedItemsRepo) client(ctx context.Context) *datastore.Client {
	var err error
	r.dclient, err = datastore.NewClient(ctx, "nyt-reading-list")
	if err != nil {
		log.Criticalf(ctx, "cannot init datastore client: %s", err)
		panic(err)
	}
	return r.dclient
}

// Get will attempt to query the underlying Datastore database for saved items
// for a single user.
func (r *DatastoreSavedItemsRepo) Get(ctx context.Context, userID string) ([]*SavedItem, error) {
	query := datastore.NewQuery(r.kind).
		Filter("UserID =", userID).
		Order("-Timestamp").
		Limit(10)

	log.Debugf(ctx, "query: %#v", query)

	var items []*SavedItem
	for iter := r.client(ctx).Run(ctx, query); ; {
		var si SavedItem
		_, err := iter.Next(&si)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		items = append(items, &si)
	}

	return items, nil
}

// Put will attempt to insert a new saved item for the user.
func (r *DatastoreSavedItemsRepo) Put(ctx context.Context, userID, url string) error {
	_, err := r.client(ctx).Put(ctx, datastore.NewKey(ctx, r.kind, userID+url, 0, nil), &SavedItem{userID, url, time.Now().UTC()})
	return err
}

// Delete will attempt to remove an item from a user's saved items.
func (r *DatastoreSavedItemsRepo) Delete(ctx context.Context, userID, url string) error {
	return r.client(ctx).Delete(ctx, datastore.NewKey(ctx, r.kind, userID+url, 0, nil))
}
