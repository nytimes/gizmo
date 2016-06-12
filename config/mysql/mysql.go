package mysql

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/NYTimes/gizmo/config"
)

// Config holds everything you need to
// connect and interact with a MySQL DB.
type Config struct {
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
	// MaxOpenConns will be used to set a MySQL
	// drivers MaxOpenConns value.
	MaxOpenConns = 1
	// MaxIdleConns will be used to set a MySQL
	// drivers MaxIdleConns value.
	MaxIdleConns = 1
)

// DB will attempt to open a sql connection with
// the credentials and the current MySQLMaxOpenConns
// and MySQLMaxIdleConns values.
// Users must import a mysql driver in their
// main to use this.
func (m *Config) DB() (*sql.DB, error) {
	db, err := sql.Open("mysql", m.String())
	if err != nil {
		return db, err
	}
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetMaxOpenConns(MaxOpenConns)
	return db, nil
}

// String will return the MySQL connection string.
func (m *Config) String() string {
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

// LoadConfigFromEnv will attempt to load a MySQL object
// from environment variables. If not populated, nil
// is returned.
func LoadConfigFromEnv() *Config {
	var mysql Config
	config.LoadEnvConfig(&mysql)
	if mysql.Host != "" {
		return &mysql
	}
	return nil
}
