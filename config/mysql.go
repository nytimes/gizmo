package config

import (
	"database/sql"
	"fmt"
	"net/url"
)

// MySQL holds everything you need to
// connect and interact with a MySQL DB.
type MySQL struct {
	User     string `envconfig:"MYSQL_USER"`
	Pw       string `envconfig:"MYSQL_PW"`
	Host     string `envconfig:"MYSQL_HOST_NAME"`
	DBName   string `envconfig:"MYSQL_DB_NAME"`
	Location string `envconfig:"MYSQL_LOCATION"`
}

// DefaultLocation is used for MySQL connections.
const DefaultLocation = "America/New_York"

var (
	// MySQLMaxOpenConns will be used to set a MySQL
	// drivers MaxOpenConns value.
	MySQLMaxOpenConns = 1
	// MySQLMaxIdleConns will be used to set a MySQL
	// drivers MaxIdleConns value.
	MySQLMaxIdleConns = 1
)

// DB will attempt to open a sql connection with
// the credentials and the current MySQLMaxOpenConns
// and MySQLMaxIdleConns values.
// Users must import a mysql driver in their
// main to use this.
func (m *MySQL) DB() (*sql.DB, error) {
	db, err := sql.Open("mysql", m.String())
	if err != nil {
		return db, err
	}
	db.SetMaxIdleConns(MySQLMaxIdleConns)
	db.SetMaxOpenConns(MySQLMaxOpenConns)
	return db, nil
}

// String will return the MySQL connection string.
func (m *MySQL) String() string {

	if m.Location != "" {
		m.Location = url.QueryEscape(DefaultLocation)
	} else {
		m.Location = url.QueryEscape(m.Location)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?loc=%s&parseTime=true",
		m.User,
		m.Pw,
		m.Host,
		m.DBName,
		m.Location,
	)
}

// LoadMySQLFromEnv will attempt to load a MySQL object
// from environment variables. If not populated, nil
// is returned.
func LoadMySQLFromEnv() *MySQL {
	var mysql MySQL
	LoadEnvConfig(&mysql)
	if mysql.Host != "" {
		return &mysql
	}
	return nil
}
