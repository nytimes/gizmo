package cookie

import "github.com/NYTimes/gizmo/config"

// Config holds information for creating
// a secure cookie.
type Config struct {
	HashKey  string `envconfig:"COOKIE_HASH_KEY"`
	BlockKey string `envconfig:"COOKIE_BLOCK_KEY"`
	Domain   string `envconfig:"COOKIE_DOMAIN"`
	Name     string `envconfig:"COOKIE_NAME"`
}

// LoadConfigFromEnv will attempt to load an Cookie object
// from environment variables. If not populated, nil
// is returned.
func LoadConfigFromEnv() Config {
	var cookie Config
	config.LoadEnvConfig(&cookie)
	return cookie
}
