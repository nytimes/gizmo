package readinglist

import (
	"context"
	"os"
	"strings"

	"cloud.google.com/go/datastore"

	"github.com/nytimes/gizmo/server/kit"
	"github.com/pkg/errors"
)

type DB interface {
	GetLinks(ctx context.Context, userID string, limit int) ([]string, error)
	PutLink(ctx context.Context, userID string, url string) error
	DeleteLink(ctx context.Context, userID string, url string) error
}

type Datastore struct {
	client *datastore.Client
}

const LinkKind = "Link"

func NewDB() (*Datastore, error) {
	ctx := context.Background()
	pid := os.Getenv("GCP_PROJECT_ID")

	ds, err := datastore.NewClient(ctx, pid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init datastore client")
	}

	return &Datastore{
		client: ds,
	}, nil
}

type linkData struct {
	UserID string
	URL    string `datastore:",noindex"`
}

func newKey(ctx context.Context, userID, url string) *datastore.Key {
	skey := strings.TrimPrefix(url, "https://www.nytimes.com/")
	return datastore.NameKey(LinkKind, reverse(userID)+"-"+skey, nil)
}

func (d *Datastore) GetLinks(ctx context.Context, userID string, limit int) ([]string, error) {
	var datas []*linkData
	q := datastore.NewQuery(LinkKind).Filter("UserID =", userID).Limit(limit)
	_, err := d.client.GetAll(ctx, q, &datas)
	links := make([]string, len(datas))
	for i, d := range datas {
		links[i] = d.URL
	}
	return links, errors.WithMessage(err, "unable to query links")
}

func (d *Datastore) DeleteLink(ctx context.Context, userID string, url string) error {
	err := d.client.Delete(ctx, newKey(ctx, userID, url))
	return errors.Wrap(err, "unable to delete url")
}

func (d *Datastore) PutLink(ctx context.Context, userID string, url string) error {
	key := newKey(ctx, userID, url)

	// run in transaction to avoid any dupes
	_, err := d.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var existing linkData
		err := tx.Get(key, &existing)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return errors.Wrap(err, "unable to check if link already exists")
		}
		// link already exists, just return
		if err != datastore.ErrNoSuchEntity {
			return nil
		}

		kit.LogMsg(ctx, userID+"---"+url)

		// put new link
		_, err = tx.Put(key, &linkData{
			UserID: userID,
			URL:    url,
		})
		return err
	})
	return errors.Wrap(err, "unable to put link")
}

// Using this to turn keys 12345, 12346 into 54321, 64321 which are easier for
// Datastore/BigTable to shard.
//
// More info: https://cloud.google.com/bigtable/docs/schema-design#row_keys_to_avoid & "Sequential numeric IDs"
func reverse(id string) string {
	runes := []rune(id)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}
