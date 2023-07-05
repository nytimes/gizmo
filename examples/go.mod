module github.com/NYTimes/gizmo/examples

replace github.com/NYTimes/gizmo => ../

require (
	cloud.google.com/go/datastore v1.10.0
	cloud.google.com/go/profiler v0.3.1 // indirect
	github.com/NYTimes/gizmo v1.2.1
	github.com/NYTimes/gziphandler v1.1.0
	github.com/NYTimes/logrotate v1.0.0
	github.com/NYTimes/sqliface v0.0.0-20180310195202-f8e6c8b78d37
	github.com/go-kit/kit v0.9.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/net v0.5.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
	google.golang.org/grpc v1.53.0
)

go 1.13
