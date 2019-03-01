module github.com/NYTimes/gizmo/examples

replace github.com/NYTimes/gizmo => ../

require (
	cloud.google.com/go v0.36.0
	github.com/NYTimes/gizmo v1.2.1
	github.com/NYTimes/gziphandler v1.1.0
	github.com/NYTimes/logrotate v1.0.0
	github.com/NYTimes/sqliface v0.0.0-20180310195202-f8e6c8b78d37
	github.com/go-kit/kit v0.8.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.2.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	golang.org/x/net v0.0.0-20190225153610-fe579d43d832
	google.golang.org/genproto v0.0.0-20190219182410-082222b4a5c5
	google.golang.org/grpc v1.18.0
)
