package observe

import (
	"context"

	"cloud.google.com/go/profiler"
	traceapi "cloud.google.com/go/trace/apiv2"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// RegisterAndObserveGCP will initiate and register Stackdriver profiling and tracing and
// metrics in environments that pass the tests in the IsGCPEnabled function. All
// exporters will be registered using the information returned by the GetServiceInfo
// function. Tracing and metrics are enabled via OpenCensus exporters. See the OpenCensus
// documentation for instructions for registering additional spans and metrics.
func RegisterAndObserveGCP(onError func(error)) error {
	if SkipObserve() {
		return nil
	}
	if !IsGCPEnabled() {
		return errors.New("Stackdriver opencensus exporter is not enabled. No observe tools will be run")
	}

	projectID, svcName, svcVersion := GetServiceInfo()

	exp, err := NewStackdriverExporter(projectID, onError)
	if err != nil {
		return errors.Wrap(err, "unable to initiate Stackdriver's opencensus exporter")
	}

	trace.RegisterExporter(exp)
	view.RegisterExporter(exp)

	err = profiler.Start(profiler.Config{
		ProjectID:      projectID,
		Service:        svcName,
		ServiceVersion: svcVersion,
	})
	return errors.Wrap(err, "unable to initiate GCP profiling client")
}

// NewStackdriverExporter will return the tracing and metrics through
// the stack driver exporter, if exists in the underlying platform.
// If exporter is registered, it returns the exporter so you can register
// it and ensure to call Flush on termination.
func NewStackdriverExporter(projectID string, onErr func(error)) (*stackdriver.Exporter, error) {
	var mr monitoredresource.Interface

	canExport := IsGAE() || IsCloudRun()
	if m := monitoredresource.Autodetect(); m != nil {
		mr = m
		canExport = true
	}
	if !canExport {
		return nil, errors.New("unable to initiate Stackdriver exporter. Environment not supported")
	}

	_, svcName, svcVersion := GetServiceInfo()
	opts, err := getSDOpts(projectID, svcName, svcVersion, mr, onErr)
	if opts == nil {
		return nil, err
	}
	return stackdriver.NewExporter(*opts)
}

// getSDOpts returns Stack Driver Options that you can pass directly
// to the OpenCensus exporter or other libraries.
func getSDOpts(projectID, service, version string, mr monitoredresource.Interface, onErr func(err error)) (*stackdriver.Options, error) {

	// this is so that you can export views from your local server up to SD if you wish
	creds, err := google.FindDefaultCredentials(context.Background(), traceapi.DefaultAuthScopes()...)
	if err != nil {
		return nil, errors.New("unable to find credentials")
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
	}, nil
}
