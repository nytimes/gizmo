package cookie

import "github.com/NYTimes/gizmo/config"

// Cookie holds information for creating
// a secure cookie.
type Cookie struct {
	HashKey  string `envconfig:"COOKIE_HASH_KEY"`
	BlockKey string `envconfig:"COOKIE_BLOCK_KEY"`
	Domain   string `envconfig:"COOKIE_DOMAIN"`
	Name     string `envconfig:"COOKIE_NAME"`
}

// LoadFromEnv will attempt to load an Cookie object
// from environment variables. If not populated, nil
// is returned.
func LoadFromEnv() Cookie {
	var cookie Cookie
	config.LoadEnvConfig(&cookie)
	return cookie
}
