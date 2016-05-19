package config

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

// MetricsType acts as an 'enum' type to represent
// the available metrics providers
type MetricsType string

const (
	// Statsd is used by config to indicate use of the StatsdProvider.
	Statsd MetricsType = "statsd"
	// DogStatsd is used by config to indicate use of the DogstatsdProvider.
	DogStatsd MetricsType = "dogstatsd"
	// Prometheus is used by config to indicate use of the PrometheusProvider.
	Prometheus MetricsType = "prometheus"
	// Graphite is used by config to indicate use of the GraphiteProvider.
	Graphite MetricsType = "graphite"
	// Expvar is used by config to indicate use of the ExpvarProvider.
	Expvar MetricsType = "expvar"
)

// Metrics config can be used to configure and instantiate a new
// go-kit/kit/metrics/provider.Provider.
type Metrics struct {
	// if empty, will server default to "expvar"
	Type MetricsType `envconfig:"METRICS_TYPE"`

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
		return nopMetricsProvider{}, nil
	}
}

type nopMetricsProvider struct{}

func (p nopMetricsProvider) NewCounter(name string, help string) metrics.Counter {
	return nopCounter{}
}
func (p nopMetricsProvider) NewHistogram(name string, help string, min int64, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	return nopHistogram{}, nil
}
func (p nopMetricsProvider) NewGauge(name string, help string) metrics.Gauge {
	return nopGauge{}
}
func (p nopMetricsProvider) Stop() {}

type nopHistogram struct{}

func (h nopHistogram) Name() string                         { return "" }
func (h nopHistogram) With(metrics.Field) metrics.Histogram { return h }
func (h nopHistogram) Observe(value int64)                  {}
func (h nopHistogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	return []metrics.Bucket{}, []metrics.Quantile{}
}

type nopGauge struct{}

func (g nopGauge) Name() string                     { return "" }
func (g nopGauge) With(metrics.Field) metrics.Gauge { return g }
func (g nopGauge) Set(value float64)                {}
func (g nopGauge) Add(delta float64)                {}
func (g nopGauge) Get() float64 {
	return 0
}

type nopCounter struct{}

func (c nopCounter) Name() string                       { return "" }
func (c nopCounter) With(metrics.Field) metrics.Counter { return c }
func (c nopCounter) Add(delta uint64)                   {}
