## `cli-sns-pub` 
* example of a simple CLI utility that emits two messages to Google Cloud Plattforn via a `gizmo/pubsub/GCPMultiPublisher`.

### The config in this example is loaded via environment variables and it utilizes the default `gizmo/config.Config`. Before running, fill in some GCP credentials and `source` the local '.env' file.