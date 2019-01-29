package kit

import (
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"github.com/go-kit/kit/log"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func gaeSDExporterOptions(projectID, service, version string, lg log.Logger) stackdriver.Options {
	return stackdriver.Options{
		ProjectID: projectID,
		MonitoredResource: mrInterface{
			typ: "gae_app",
			labels: map[string]string{
				"project_id": projectID,
			},
		},
		OnError: func(err error) {
			lg.Log("error", err,
				"message", "tracing client encountered an error")
		},
		DefaultMonitoringLabels: &stackdriver.Labels{},
		DefaultTraceAttributes: map[string]interface{}{
			"service": service,
			"version": version,
		},
	}
}

func googleProjectID() string {
	return os.Getenv("GOOGLE_CLOUD_PROJECT")
}

func isGKE() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}

func gkeSDExporterOptions(projectID string, lg log.Logger) stackdriver.Options {
	return stackdriver.Options{
		ProjectID:         projectID,
		MonitoredResource: monitoredresource.Autodetect(),
		OnError: func(err error) {
			lg.Log("error", err,
				"message", "tracing client encountered an error")
		},
		DefaultMonitoringLabels: &stackdriver.Labels{},
	}
}

func initSDExporter(opt stackdriver.Options) error {
	exporter, err := stackdriver.NewExporter(opt)
	if err != nil {
		return err
	}
	trace.RegisterExporter(exporter)
	view.RegisterExporter(exporter)
	return nil
}

// implements contrib.go.opencensus.io/exporter/stackdriver/monitoredresource.Interface
type mrInterface struct {
	typ    string
	labels map[string]string
}

func (g mrInterface) MonitoredResource() (string, map[string]string) {
	return g.typ, g.labels
}
