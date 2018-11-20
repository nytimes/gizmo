package kit

import (
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/go-kit/kit/log"
	"go.opencensus.io/trace"
)

func initGAETrace(projectID, service, version string, lg log.Logger) error {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
		MonitoredResource: gaeInterface{
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
			"g.co/gae/app/module_version": version,
			"service":                     service,
			"version":                     version,
		},
	})
	if err != nil {
		return err
	}
	trace.RegisterExporter(exporter)
	return nil
}

// implements contrib.go.opencensus.io/exporter/stackdriver/monitoredresource.Interface
type gaeInterface struct {
	labels map[string]string
}

func (g gaeInterface) MonitoredResource() (string, map[string]string) {
	return "gae_app", g.labels
}
