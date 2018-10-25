# This is an experimental package in Gizmo!

* The rationale behind this package:
    * A more opinionated server with fewer choices.
    * go-kit is used for serving HTTP/JSON & gRPC is used for serving HTTP2/RPC
    * Logs always go to stdout/stderr by default, but if running on App Engine, trace enabled Stackdriver logging will be used instead.
    * Using Go's 1.8 graceful HTTP shutdown
    * Services using this package are meant for deploy to GCP.

* If you experience any issues please create an issue and/or reach out on the #gizmo channel with what you've found.
