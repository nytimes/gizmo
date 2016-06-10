package mysql

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/NYTimes/gizmo/config"
)

// MySQL holds everything you need to
// connect and interact with a MySQL DB.
type MySQL struct {
	User     string `envconfig:"MYSQL_USER"`
	Pw       string `envconfig:"MYSQL_PW"`
	Host     string `envconfig:"MYSQL_HOST_NAME"`
	Port     int    `envconfig:"MYSQL_PORT"`
	DBName   string `envconfig:"MYSQL_DB_NAME"`
	Location string `envconfig:"MYSQL_LOCATION"`
}

const (
	// DefaultLocation is the default location for MySQL connections.
	DefaultLocation = "America/New_York"
	// DefaultMySQLPort is the default port for MySQL connections.
	DefaultMySQLPort = 3306
)

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
	if m.Port == 0 {
		m.Port = DefaultMySQLPort
	}

	if m.Location != "" {
		m.Location = url.QueryEscape(m.Location)
	} else {
		m.Location = url.QueryEscape(DefaultLocation)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?loc=%s&parseTime=true",
		m.User,
		m.Pw,
		m.Host,
		m.Port,
		m.DBName,
		m.Location,
	)
}

// LoadFromEnv will attempt to load a MySQL object
// from environment variables. If not populated, nil
// is returned.
func LoadFromEnv() *MySQL {
	var mysql MySQL
	config.LoadEnvConfig(&mysql)
	if mysql.Host != "" {
		return &mysql
	}
	return nil
}
