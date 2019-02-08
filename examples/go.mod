module github.com/NYTimes/gizmo/examples

require (
	cloud.google.com/go v0.36.0
	github.com/NYTimes/gizmo v1.1.0
	github.com/NYTimes/gziphandler v1.0.1
	github.com/NYTimes/logrotate v0.0.0-20170824154650-2b6e866fd507
	github.com/NYTimes/sqliface v0.0.0-20180310195202-f8e6c8b78d37
	github.com/go-kit/kit v0.8.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.2.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	golang.org/x/net v0.0.0-20190206173232-65e2d4e15006
	google.golang.org/genproto v0.0.0-20190201180003-4b09977fb922
	google.golang.org/grpc v1.18.0
)

replace github.com/NYTimes/gizmo => ../../gizmo
