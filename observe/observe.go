// Package observe provides functions
// that help with setting tracing/metrics
// in cloud providers, mainly GCP.
package observe // import "github.com/NYTimes/gizmo/observe"

import (
	"os"

	"cloud.google.com/go/profiler"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
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
		return errors.New("environment is not GCP enabled, no observe tools will be run")
	}

	projectID, svcName, svcVersion := GetServiceInfo()

	exp, err := NewStackdriverExporter(projectID, onError)
	if err != nil {
		return errors.Wrap(err, "unable to initiate error tracing exporter")
	}
	trace.RegisterExporter(exp)
	view.RegisterExporter(exp)

	err = profiler.Start(profiler.Config{
		ProjectID:      projectID,
		Service:        svcName,
		ServiceVersion: svcVersion,
	})
	return errors.Wrap(err, "unable to initiate profiling client")
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

// GetGAEInfo returns the service and the version of the
// GAE application.
func GetGAEInfo() (service, version string) {
	return os.Getenv("GAE_SERVICE"), os.Getenv("GAE_VERSION")
}

// IsCloudRun tells you whether your program is running
// within the Cloud Run platform.
func IsCloudRun() bool {
	return os.Getenv("K_CONFIGURATION") != ""
}

// GetCloudRunInfo returns the service and the version of the
// Cloud Run application.
func GetCloudRunInfo() (service, version, config string) {
	return os.Getenv("K_SERVICE"), os.Getenv("K_REVISION"), os.Getenv("K_CONFIGURATION")
}

// GetServiceInfo returns the GCP Project ID,
// the service name and version (GAE or through
// SERVICE_NAME/SERVICE_VERSION env vars). Note
// that SERVICE_NAME/SERVICE_VERSION are not standard but
// your application can pass them in as variables
// to be included in your trace attributes
func GetServiceInfo() (projectID, service, version string) {
	switch {
	case IsGAE():
		service, version = GetGAEInfo()
	case IsCloudRun():
		service, version, _ = GetCloudRunInfo()
	default:
		service, version = os.Getenv("SERVICE_NAME"), os.Getenv("SERVICE_VERSION")
	}
	return GoogleProjectID(), service, version
}

// IsGCPEnabled returns whether the running application
// is inside GCP or has access to its products.
func IsGCPEnabled() bool {
	return IsGAE() || IsCloudRun() || monitoredresource.Autodetect() != nil
}

// IsDatadogEnabled checks if exporitng metrics and traces to Datadog should enabled (by
// setting DATADOG_ENABLED environemnt variable) and if the Datadog's agent address is
// provided (by DATADOG_ADDR environsmnt variable)
func IsDatadogEnabled() bool {
	return os.Getenv("DATADOG_ENABLED") != "" && getDatadogAddr() != ""
}

// SkipObserve checks if the GIZMO_SKIP_OBSERVE environment variable has been populated.
// This may be used along with local development to cut down on long startup times caused
// by the 'monitoredresource.Autodetect()' call in IsGCPEnabled().
func SkipObserve() bool {
	return os.Getenv("GIZMO_SKIP_OBSERVE") != ""
}
