package observe

import (
	"context"

	traceapi "cloud.google.com/go/trace/apiv2"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// NewStackdriverExporter will return the tracing and metrics through
// the stack driver exporter, if exists in the underlying platform.
// If exporter is registered, it returns the exporter so you can register
// it and ensure to call Flush on termination.
func NewStackdriverExporter(projectID string, onErr func(error)) (*stackdriver.Exporter, error) {
	_, svcName, svcVersion := GetServiceInfo()
	opts := getSDOpts(projectID, svcName, svcVersion, onErr)
	if opts == nil {
		return nil, nil
	}
	return stackdriver.NewExporter(*opts)
}

// getSDOpts returns Stack Driver Options that you can pass directly
// to the OpenCensus exporter or other libraries.
func getSDOpts(projectID, service, version string, onErr func(err error)) *stackdriver.Options {
	var mr monitoredresource.Interface

	// this is so that you can export views from your local server up to SD if you wish
	creds, err := google.FindDefaultCredentials(context.Background(), traceapi.DefaultAuthScopes()...)
	if err != nil {
		return nil
	}
	canExport := IsGAE() || IsCloudRun()
	if m := monitoredresource.Autodetect(); m != nil {
		mr = m
		canExport = true
	}
	if !canExport {
		return nil
	}

	return &stackdriver.Options{
		ProjectID:         projectID,
		MonitoredResource: mr,
		MonitoringClientOptions: []option.ClientOption{
			option.WithCredentials(creds),
		},
		TraceClientOptions: []option.ClientOption{
			option.WithCredentials(creds),
		},
		OnError: onErr,
		DefaultTraceAttributes: map[string]interface{}{
			"service": service,
			"version": version,
		},
	}
}
