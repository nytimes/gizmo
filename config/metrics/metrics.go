package metrics

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/provider"

	"github.com/NYTimes/gizmo/config"
)

// Type acts as an 'enum' type to represent
// the available metrics providers
type Type string

const (
	// Statsd is used by config to indicate use of the statsdProvider.
	Statsd Type = "statsd"
	// DogStatsd is used by config to indicate use of the dogstatsdProvider.
	DogStatsd Type = "dogstatsd"
	// Prometheus is used by config to indicate use of the prometheusProvider.
	Prometheus Type = "prometheus"
	// Graphite is used by config to indicate use of the graphiteProvider.
	Graphite Type = "graphite"
	// Expvar is used by config to indicate use of the expvarProvider.
	Expvar Type = "expvar"
	// Discard is used by config to indicate use of the discardProvider.
	Discard Type = "discard"
)

// Metrics config can be used to configure and instantiate a new
// go-kit/kit/metrics/provider.Provider.
type Metrics struct {
	Type Type `envconfig:"METRICS_TYPE"`

	// Prefix will be prefixed onto
	// any metric name.
	Prefix string `envconfig:"METRICS_PREFIX"`

	// Namespace is used by prometheus.
	Namespace string `envconfig:"METRICS_NAMESPACE"`
	// Subsystem is used by prometheus.
	Subsystem string `envconfig:"METRICS_SUBSYSTEM"`

	// Used by statsd, graphite and dogstatsd
	Interval time.Duration `envconfig:"METRICS_INTERVAL"`

	// Used by statsd, graphite and dogstatsd.
	Addr string `envconfig:"METRICS_ADDR"`
	// Used by statsd, graphite and dogstatsd to dial a connection.
	// If empty, will default to "udp".
	Network string `envconfig:"METRICS_NETWORK"`

	// Used by expvar only.
	// if empty, will default to "/debug/vars"
	Path string `envconfig:"METRICS_PATH"`

	// Used by graphite only.
	// If none provided, kit/log/NewNopLogger will be used.
	Logger log.Logger
}

// LoadFromEnv will attempt to load a Metrics object
// from environment variables.
func LoadFromEnv() Metrics {
	var mets Metrics
	config.LoadEnvConfig(&mets)
	return mets
}

// NewProvider will use the values in the Metrics config object
// to generate a new go-kit/metrics/provider.Provider implementation.
// If no type is given, a no-op implementation will be used.
func (cfg Metrics) NewProvider() (provider.Provider, error) {
	if cfg.Logger == nil {
		cfg.Logger = log.NewNopLogger()
	}
	if cfg.Path == "" {
		cfg.Path = "/debug/vars"
	}
	if cfg.Interval == 0 {
		cfg.Interval = time.Second * 30
	}
	switch cfg.Type {
	case Statsd:
		return provider.NewStatsdProvider(cfg.Network, cfg.Addr,
			cfg.Prefix, cfg.Interval, cfg.Logger)
	case DogStatsd:
		return provider.NewDogStatsdProvider(cfg.Network, cfg.Addr,
			cfg.Prefix, cfg.Interval, cfg.Logger)
	case Graphite:
		return provider.NewGraphiteProvider(cfg.Network, cfg.Addr,
			cfg.Prefix, cfg.Interval, cfg.Logger)
	case Prometheus:
		return provider.NewPrometheusProvider(cfg.Namespace, cfg.Subsystem), nil
	case Expvar:
		return provider.NewExpvarProvider(cfg.Prefix), nil
	default:
		return provider.NewDiscardProvider(), nil
	}
}
