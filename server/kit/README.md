# This is an experimental package in Gizmo!

* The rationale behind this package:
    * A more opinionated server with fewer choices.
    * go-kit is used for serving HTTP/JSON & gRPC is used for serving HTTP2/RPC
    * Monitoring and metrics are handled by a sidecar (ie. Cloud Endpoints)
    * Logs always go to stdout/stderr
    * Using Go's 1.8 graceful HTTP shutdown
    * Services using this package are meant for deploy to GCP with GKE and Cloud Endpoints.

* If you experience any issues please create an issue and/or reach out on the #gizmo channel with what you've found 
