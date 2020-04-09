package observe

import (
	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"github.com/pkg/errors"
)

// DatadogExporterConfig provides configuration for the Datadog exporter.
// Config can be initiated with envconfig from environment
type DatadogExporterConfig struct {
	DatadogExporterEnabled        bool   `default:"false" split_words:"true"`
	DatadogExporterMetricsAddress string `split_words:"true"`
	DatadogExporterTracesAddress  string `split_words:"true"`
	DatadogExporterNamespace      string `default:"opencensus" split_words:"true"`
}

// NewDatadogExporter will return Datadog's opencensus exporter if it's enabled (through
// DATADOG_EXPORTER_ENABLED env variable). When the exporter is disabled, it will
// return nil. Exporter will send metrics and traces to Datadog's agent using
// addresses specified through DatadogExporterConfig
func NewDatadogExporter(config DatadogExporterConfig, onErr func(error)) (*datadog.Exporter, error) {

	if config.DatadogExporterMetricsAddress == "" && config.DatadogExporterTracesAddress == "" {
		return nil, errors.New("Missing Datadog agent's address for metrics and traces")
	}

	_, service, version := GetServiceInfo()

	opts := getDatadogOpts(config, service, version, onErr)

	return datadog.NewExporter(opts)
}

// getDatadogOpts returns Datadog Options that you can pass directly
// to the OpenCensus exporter or other libraries.
func getDatadogOpts(config DatadogExporterConfig, service, version string, onErr func(err error)) datadog.Options {

	return datadog.Options{
		// Namespace specifies the namespaces to which metric keys are appended.
		// TODO: Figure out what the namespace should be. Can be either a projectID or something else.
		Namespace: config.DatadogExporterNamespace,

		// Service specifies the service name used for tracing.
		Service: service,

		// TraceAddr specifies the host[:port] address of the Datadog Trace Agent.
		// It defaults to localhost:8126.
		TraceAddr: config.DatadogExporterTracesAddress,

		// StatsAddr specifies the host[:port] address for DogStatsD. It defaults
		// to localhost:8125.
		StatsAddr: config.DatadogExporterMetricsAddress,

		// OnError specifies a function that will be called if an error occurs during
		// processing stats or metrics.
		OnError: onErr,

		// // Tags specifies a set of global tags to attach to each metric.
		// Tags []string

		// GlobalTags holds a set of tags that will automatically be applied to all
		// exported spans.
		GlobalTags: map[string]interface{}{
			"service": service,
			"version": version,
		},
		// // DisableCountPerBuckets specifies whether to emit count_per_bucket metrics
		// DisableCountPerBuckets bool
	}
}
