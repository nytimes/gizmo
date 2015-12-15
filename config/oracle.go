package config

import (
	"database/sql"
	"fmt"
)

// Oracle holds everything you need to
// connect and interact with an Oracle DB.
type Oracle struct {
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
func (o *Oracle) DB() (*sql.DB, error) {
	return sql.Open("oci8", o.String())
}

// String will return the Oracle connection string.
func (o *Oracle) String() string {
	if len(o.ConnectString) > 0 {
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

// LoadOracleFromEnv will attempt to load an OracleCreds object
// from environment variables. If not populated, nil
// is returned.
func LoadOracleFromEnv() *Oracle {
	var ora Oracle
	LoadEnvConfig(&ora)
	if ora.Host == "" {
		return nil
	}
	return &ora
}
