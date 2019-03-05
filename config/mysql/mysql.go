package mysql // import "github.com/NYTimes/gizmo/config/mysql"

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/kelseyhightower/envconfig"
)

// Config holds everything you need to
// connect and interact with a MySQL DB.
type Config struct {
	Pw              string `envconfig:"MYSQL_PW"`
	User            string `envconfig:"MYSQL_USER"`
	Port            int    `envconfig:"MYSQL_PORT"`
	DBName          string `envconfig:"MYSQL_DB_NAME"`
	Location        string `envconfig:"MYSQL_LOCATION"`
	Host            string `envconfig:"MYSQL_HOST_NAME"`
	ReadTimeout     string `envconfig:"MYSQL_READ_TIMEOUT"`
	WriteTimeout    string `envconfig:"MYSQL_WRITE_TIMEOUT"`
	AddtlDSNOptions string `envconfig:"MYSQL_ADDTL_DSN_OPTIONS"`
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
	var port int
	if m.Port == 0 {
		port = DefaultMySQLPort
	} else {
		port = m.Port
	}

	var location string
	if m.Location != "" {
		location = url.QueryEscape(m.Location)
	} else {
		location = url.QueryEscape(DefaultLocation)
	}

	args, _ := url.ParseQuery(m.AddtlDSNOptions)

	args.Set("parseTime", "true")

	if m.ReadTimeout != "" {
		args.Set("readTimeout", m.ReadTimeout)
	}
	if m.WriteTimeout != "" {
		args.Set("writeTimeout", m.WriteTimeout)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?loc=%s&%s",
		m.User,
		m.Pw,
		m.Host,
		port,
		m.DBName,
		location,
		args.Encode(),
	)
}

// LoadConfigFromEnv will attempt to load a MySQL object
// from environment variables. If not populated, nil
// is returned.
func LoadConfigFromEnv() *Config {
	var mysql Config
	envconfig.Process("", &mysql)
	if mysql.Host != "" {
		return &mysql
	}
	return nil
}
