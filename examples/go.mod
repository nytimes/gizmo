module github.com/NYTimes/gizmo/examples

replace github.com/NYTimes/gizmo => ../

require (
	cloud.google.com/go v0.57.0
	github.com/NYTimes/gizmo v1.2.1
	github.com/NYTimes/gziphandler v1.1.0
	github.com/NYTimes/logrotate v1.0.0
	github.com/NYTimes/sqliface v0.0.0-20180310195202-f8e6c8b78d37
	github.com/go-kit/kit v0.9.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/net v0.0.0-20200501053045-e0ff5e5a1de5
	google.golang.org/genproto v0.0.0-20200430143042-b979b6f78d84
	google.golang.org/grpc v1.29.1
)

go 1.13
