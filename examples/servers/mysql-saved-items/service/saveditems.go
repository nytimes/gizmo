package service

import (
	"database/sql"
	"time"

	"github.com/NYTimes/gizmo/config/mysql"
	"github.com/NYTimes/sqliface"
)

type (
	// SavedItemsRepo is an interface layer between
	// our service and our database. Abstracting these methods
	// out of a pure implementation helps with testing.
	SavedItemsRepo interface {
		Get(uint64) ([]*SavedItem, error)
		Put(uint64, string) error
		Delete(uint64, string) error
	}

	// MySQLSavedItemsRepo is an implementation of the repo
	// interface built on top of MySQL.
	MySQLSavedItemsRepo struct {
		db *sql.DB
	}

	// SavedItem represents an article, blog, interactive, etc.
	// that a user wants to save for reading later.
	SavedItem struct {
		UserID    uint64    `json:"user_id"`
		URL       string    `json:"url"`
		Timestamp time.Time `json:"timestamp"`
	}
)

// NewSavedItemsRepo will attempt to connect to to MySQL and
// return a SavedItemsRepo implementation.
func NewSavedItemsRepo(cfg *mysql.Config) (SavedItemsRepo, error) {
	db, err := cfg.DB()
	if err != nil {
		return nil, err
	}
	return &MySQLSavedItemsRepo{db}, nil
}

// Get will attempt to query the underlying MySQL database for saved items
// for a single user.
func (r *MySQLSavedItemsRepo) Get(userID uint64) ([]*SavedItem, error) {
	query := `SELECT
				user_id,
				url,
				timestamp
			FROM saved_items
			WHERE user_id = ?
			ORDER BY timestamp DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

func scanItems(rows sqliface.Rows) ([]*SavedItem, error) {
	var err error
	// initializing so we return an empty array in case of 0
	items := []*SavedItem{}
	for rows.Next() {
		item := &SavedItem{}
		err = rows.Scan(&item.UserID, &item.URL, &item.Timestamp)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// Put will attempt to insert a new saved item for the user.
func (r *MySQLSavedItemsRepo) Put(userID uint64, url string) error {
	query := `INSERT INTO saved_items (user_id, url, timestamp)
				VALUES (?, ?, NOW())
			  ON DUPLICATE KEY UPDATE timestamp = NOW()`
	_, err := r.db.Exec(query, userID, url)
	return err
}

// Delete will attempt to remove an item from a user's saved items.
func (r *MySQLSavedItemsRepo) Delete(userID uint64, url string) error {
	query := `DELETE FROM saved_items
			  WHERE user_id = ? AND url = ?`
	_, err := r.db.Exec(query, userID, url)
	return err
}
