// Package observe provides functions
// that help with setting tracing/metrics
// in cloud providers, mainly GCP.
package observe

import (
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// NewStackDriverExporter will return the tracing and metrics through
// the stack driver exporter, if exists in the underlying platform.
// If exporter is registered, it returns the exporter so you can register
// it and ensure to call Flush on termination.
func NewStackDriverExporter(projectID string, onErr func(error)) (*stackdriver.Exporter, error) {
	svcName, svcVersion := "", ""
	if IsGAE() {
		_, svcName, svcVersion = GetGAEInfo()
	} else if n, v := os.Getenv("SERVICE_NAME"), os.Getenv("SERVICE_VERSION"); n != "" {
		svcName, svcVersion = n, v
	}
	opts := SDExporterOptions(projectID, svcName, svcVersion, onErr)
	if opts == nil {
		return nil, nil
	}
	return stackdriver.NewExporter(*opts)
}

// NewPrometheusExporter return a prometheus Exporter for OpenCensus.
func NewPrometheusExporter(opts prometheus.Options) (*prometheus.Exporter, error) {
	return NewPrometheusExporter(opts)
}

// GoogleProjectID returns the GCP Project ID
// that can be used to instantiate various
// GCP clients such as Stack Driver.
func GoogleProjectID() string {
	return os.Getenv("GOOGLE_CLOUD_PROJECT")
}

// IsGAE tells you whether your program is running
// within the App Engine platform.
func IsGAE() bool {
	return os.Getenv("GAE_DEPLOYMENT_ID") != ""
}

// GetGAEInfo returns the GCP Project ID,
// the service, and the version of the application.
func GetGAEInfo() (projectID, service, version string) {
	return GoogleProjectID(),
		os.Getenv("GAE_SERVICE"),
		os.Getenv("GAE_VERSION")
}

// SDExporterOptions returns Stack Driver Options that you can pass directly
// to the OpenCensus exporter or other libraries.
func SDExporterOptions(projectID, service, version string, onErr func(err error)) *stackdriver.Options {
	var mr monitoredresource.Interface
	if m := monitoredresource.Autodetect(); m != nil {
		mr = m
	} else if IsGAE() {
		mr = gaeInterface{
			typ: "gae_app",
			labels: map[string]string{
				"project_id": projectID,
			},
		}
	}
	if mr == nil {
		return nil
	}

	return &stackdriver.Options{
		ProjectID:         projectID,
		MonitoredResource: mr,
		OnError:           onErr,
		DefaultTraceAttributes: map[string]interface{}{
			"service": service,
			"version": version,
		},
	}
}

// IsGCPAccessible returns whether the running application
// is inside GCP or has access to its products.
func IsGCPAccessible() bool {
	return monitoredresource.Autodetect() != nil || IsGAE()
}

// InitSDExporter will initialize the OpenCensus tracing/metrics exporter
func InitSDExporter(opts stackdriver.Options) error {
	exporter, err := stackdriver.NewExporter(opts)
	if err != nil {
		return err
	}
	trace.RegisterExporter(exporter)
	view.RegisterExporter(exporter)
	return nil
}

// implements contrib.go.opencensus.io/exporter/stackdriver/monitoredresource.Interface
type gaeInterface struct {
	typ    string
	labels map[string]string
}

func (g gaeInterface) MonitoredResource() (string, map[string]string) {
	return g.typ, g.labels
}
