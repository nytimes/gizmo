module github.com/NYTimes/gizmo/examples

replace github.com/NYTimes/gizmo => ../

require (
	cloud.google.com/go v0.38.0
	github.com/NYTimes/gizmo v1.2.1
	github.com/NYTimes/gziphandler v1.1.0
	github.com/NYTimes/logrotate v1.0.0
	github.com/NYTimes/sqliface v0.0.0-20180310195202-f8e6c8b78d37
	github.com/go-kit/kit v0.9.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610
	google.golang.org/grpc v1.22.0
)

go 1.13
