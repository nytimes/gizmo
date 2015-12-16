package config

import "flag"

// DefaultConfigLocation is the default filepath for JSON config files.
const DefaultConfigLocation = "/opt/nyt/etc/conf.json"

var (
	// LogCLI is a pointer to the value of the '-log' command line flag. It is meant to declare
	// an application logging location.
	LogCLI = flag.String("log", "", "Application log location")
	// HTTPAccessLogCLI is a pointer to the value of the '-http-access-log' command line flag. It is meant to
	// declare an access log location for HTTP services.
	HTTPAccessLogCLI = flag.String("http-access-log", "", "HTTP access log location")
	// RPCAccessLogCLI is a pointer to the value of the '-rpc-access-log' command line flag. It is meant to
	// declare an acces log location for RPC services.
	RPCAccessLogCLI = flag.String("rpc-access-log", "", "RPC access log location")
	// ConfigLocationCLI is a pointer to the value of the '-config' command line flag. It is meant to declare
	// the location of a config file. It defaults to `DefaultConfigLocation`.
	ConfigLocationCLI = flag.String("config", DefaultConfigLocation, "Application config file location")
	// HTTPPortCLI is a pointer to the value for the '-http' flag. It is meant to declare the port
	// number to serve HTTP services.
	HTTPPortCLI = flag.Int("http", 0, "Port to run an HTTP server on")
	// RPCPortCLI is a pointer to the value for the '-rpc' flag. It is meant to declare the port
	// number to serve RPC services.
	RPCPortCLI = flag.Int("rpc", 0, "Port to run an RPC server on")
)

// SetLogOverride will check `*LogCLI` for any values
// and override the given string pointer if LogCLI is set.
// If LogCLI is set to "dev", the given log var will be set to "".
func SetLogOverride(log *string) {
	// if a user passes in 'dev' log flag, override the
	// App log to signal for stderr logging.
	if *LogCLI != "" {
		*log = *LogCLI
		if *LogCLI == "dev" {
			*log = ""
		}
	}
}

// SetServerOverrides will check the *CLI variables for any values
// and override the values in the given config if they are set.
// If LogCLI is set to "dev", the given `Log` pointer will be set to an
// empty string.
func SetServerOverrides(c *Server) {
	SetLogOverride(&c.Log)

	if *HTTPAccessLogCLI != "" {
		c.HTTPAccessLog = *HTTPAccessLogCLI
	}

	if *RPCAccessLogCLI != "" {
		c.RPCAccessLog = *RPCAccessLogCLI
	}

	if *HTTPPortCLI > 0 {
		c.HTTPPort = *HTTPPortCLI
	}

	if *RPCPortCLI > 0 {
		c.RPCPort = *RPCPortCLI
	}
}
