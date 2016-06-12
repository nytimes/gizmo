package oracle

import (
	"database/sql"
	"fmt"

	"github.com/NYTimes/gizmo/config"
)

// Config holds everything you need to
// connect and interact with an Oracle DB.
type Config struct {
	User          string `envconfig:"ORACLE_USER"`
	Pw            string `envconfig:"ORACLE_PW"`
	Host          string `envconfig:"ORACLE_HOST_NAME"`
	Port          int    `envconfig:"ORACLE_PORT"`
	DBName        string `envconfig:"ORACLE_DB_NAME"`
	ConnectString string `envconfig:"ORACLE_CONNECT_STRING"`
}

// DB will attempt to open a sql connection.
// Users must import an oci8 driver in their
// main to use this.
func (o *Config) DB() (*sql.DB, error) {
	return sql.Open("oci8", o.String())
}

// String will return the Oracle connection string.
func (o *Config) String() string {
	if o.ConnectString != "" {
		return fmt.Sprintf("%s/%s@%s",
			o.User,
			o.Pw,
			o.ConnectString,
		)
	}
	return fmt.Sprintf("%s/%s@%s:%d/%v",
		o.User,
		o.Pw,
		o.Host,
		o.Port,
		o.DBName,
	)
}

// LoadConfigFromEnv will attempt to load an OracleCreds object
// from environment variables.
func LoadConfigFromEnv() Config {
	var ora Config
	config.LoadEnvConfig(&ora)
	return ora
}
