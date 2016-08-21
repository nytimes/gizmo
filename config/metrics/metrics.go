package metrics

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
	"github.com/go-kit/kit/metrics/graphite"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/go-kit/kit/metrics/statsd"

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

// Config can be used to configure and instantiate a new
// go-kit/kit/metrics/provider.Provider.
type Config struct {
	// if empty, will server default to "expvar"
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

// LoadConfigFromEnv will attempt to load a Metrics object
// from environment variables.
func LoadConfigFromEnv() Config {
	var mets Config
	config.LoadEnvConfig(&mets)
	return mets
}

// NewProvider will use the values in the Metrics config object
// to generate a new go-kit/metrics/provider.Provider implementation.
// If no type is given, a no-op implementation will be used.
func (cfg Config) NewProvider() provider.Provider {
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
		stsd := statsd.New(cfg.Prefix, cfg.Logger)
		tick := time.NewTicker(cfg.Interval)
		go stsd.SendLoop(tick.C, cfg.Network, cfg.Addr)
		return provider.NewStatsdProvider(stsd, tick.Stop)
	case DogStatsd:
		stsd := dogstatsd.New(cfg.Prefix, cfg.Logger)
		tick := time.NewTicker(cfg.Interval)
		go stsd.SendLoop(tick.C, cfg.Network, cfg.Addr)
		return provider.NewDogstatsdProvider(stsd, tick.Stop)
	case Graphite:
		grpht := graphite.New(cfg.Prefix, cfg.Logger)
		tick := time.NewTicker(cfg.Interval)
		go grpht.SendLoop(tick.C, cfg.Network, cfg.Addr)
		return provider.NewGraphiteProvider(grpht, tick.Stop)
	case Prometheus:
		return provider.NewPrometheusProvider(cfg.Namespace, cfg.Subsystem)
	case Expvar:
		return provider.NewExpvarProvider()
	default:
		return provider.NewDiscardProvider()
	}
}
