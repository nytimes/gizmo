package ocexporters

import (
	"fmt"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// NewOCExporters returns trace/client exporters
// based on the given environment. ProjectID
// is the GCP projectID for stackdriver (or the prometheus namespace) and backend
// specifies whether the backend is StackDriver or Prometheus
func NewOCExporters(projectID, backend string, onErr func(err error)) (trace.Exporter, view.Exporter, error) {
	switch backend {
	case "stackdriver":
		return getSDExporter(projectID, onErr)
	case "prometheus":
		return getPrometheusExporter(projectID, onErr)
	}

	return nil, nil, fmt.Errorf("unrecognized backend: %v", backend)
}

func getPrometheusExporter(namespace string, onErr func(error)) (trace.Exporter, view.Exporter, error) {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: namespace,
		OnError:   onErr,
	})
	return nil, pe, err
}

func getSDExporter(projectID string, onErr func(error)) (trace.Exporter, view.Exporter, error) {
	svcName, svcVersion := "", ""
	if IsGAE() {
		_, svcName, svcVersion = GetGAEInfo()
	} else if n, v := os.Getenv("SERVICE_NAME"), os.Getenv("SERVICE_VERSION"); n != "" {
		svcName, svcVersion = n, v
	}
	opts := getSDExporterOptions(projectID, svcName, svcVersion, onErr)
	if opts == nil {
		return nil, nil, nil
	}
	exp, err := stackdriver.NewExporter(*opts)
	return exp, exp, err
}

// getSDExporterOptions returns Stack Driver Options that you can pass directly
// to the OpenCensus exporter or other libraries.
func getSDExporterOptions(projectID, service, version string, onErr func(err error)) *stackdriver.Options {
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
		ProjectID:               projectID,
		MonitoredResource:       mr,
		OnError:                 onErr,
		DefaultMonitoringLabels: &stackdriver.Labels{},
		DefaultTraceAttributes: map[string]interface{}{
			"service": service,
			"version": version,
		},
	}
}

// implements contrib.go.opencensus.io/exporter/stackdriver/monitoredresource.Interface
type gaeInterface struct {
	typ    string
	labels map[string]string
}

func (g gaeInterface) MonitoredResource() (string, map[string]string) {
	return g.typ, g.labels
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
