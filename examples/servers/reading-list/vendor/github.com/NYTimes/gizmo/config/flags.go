package config

import "flag"

// DefaultConfigLocation is the default filepath for JSON config files.
const DefaultConfigLocation = "/opt/nyt/etc/conf.json"

// SetLogOverride will check `*LogCLI` for any values
// and override the given string pointer if LogCLI is set.
// If LogCLI is set to "dev", the given log var will be set to "".
func SetLogOverride(log *string) {
	// LogCLI is a pointer to the value of the '-log' command line flag. It is meant to declare
	// an application logging location.
	logCLI := flag.String("log", "", "Application log location")

	flag.Parse()

	// if a user passes in 'dev' log flag, override the
	// App log to signal for stderr logging.
	if *logCLI != "" {
		*log = *logCLI
		if *logCLI == "dev" {
			*log = ""
		}
	}
}

// SetFlagOverrides will add and check a `log` and `config` CLI flag to the
// current process, call flag.Parse() and will overwrite the passed in string
// pointer if the flag exists.
func SetFlagOverrides(log *string, config *string) {
	// create the flag
	cfg := flag.String("config", "", "Application log location")
	SetLogOverride(log)

	// if a user passes in 'dev' log flag, override the
	// App log to signal for stderr logging.
	if *cfg != "" {
		*config = *cfg
	}
}
