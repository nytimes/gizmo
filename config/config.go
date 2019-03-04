package config // import "github.com/NYTimes/gizmo/config"

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// EnvAppName is used as a prefix for environment variable
// names when using the LoadXFromEnv funcs.
// It defaults to empty.
var EnvAppName = ""

// LoadJSONFile is a helper function to read a config file into whatever
// config struct you need. For example, your custom config could be composed
// of one or more of the given Config, AWS, MySQL, Oracle or MongoDB structs.
func LoadJSONFile(fileName string, cfg interface{}) {
	cb, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read config file '%s': %s", fileName, err)
	}

	if err = json.Unmarshal(cb, &cfg); err != nil {
		log.Fatalf("Unable to parse JSON in config file '%s': %s", fileName, err)
	}
}
