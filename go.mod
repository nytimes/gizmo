module github.com/NYTimes/gizmo

require (
	cloud.google.com/go v0.32.0
	contrib.go.opencensus.io/exporter/stackdriver v0.7.0 // indirect
	git.apache.org/thrift.git v0.0.0-20181112125854-24918abba929 // indirect
	github.com/DataDog/datadog-go v0.0.0-20180822151419-281ae9f2d895 // indirect
	github.com/NYTimes/gziphandler v1.0.1
	github.com/NYTimes/logrotate v0.0.0-20170824154650-2b6e866fd507
	github.com/NYTimes/sqliface v0.0.0-20180310195202-f8e6c8b78d37
	github.com/Shopify/sarama v1.19.0
	github.com/Shopify/toxiproxy v2.1.3+incompatible // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.15.73
	github.com/bradfitz/gomemcache v0.0.0-20180710155616-bc664df96737
	github.com/circonus-labs/circonus-gometrics v2.2.4+incompatible // indirect
	github.com/circonus-labs/circonusllhist v0.1.0 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/go-ini/ini v1.39.0 // indirect
	github.com/go-kit/kit v0.8.0
	github.com/go-logfmt/logfmt v0.3.0 // indirect
	github.com/go-sql-driver/mysql v1.4.0
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/lint v0.0.0-20181026193005-c67002cb31c3 // indirect
	github.com/golang/protobuf v1.2.0
	github.com/google/go-cmp v0.2.0
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181103185306-d547d1d9531e // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/websocket v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.5.1 // indirect
	github.com/hashicorp/consul v1.3.0
	github.com/hashicorp/go-cleanhttp v0.5.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.0.0-20180718195005-e651d75abec6 // indirect
	github.com/hashicorp/go-rootcerts v0.0.0-20160503143440-6bb64b370b90 // indirect
	github.com/hashicorp/go-sockaddr v0.0.0-20180320115054-6d291a969b86 // indirect
	github.com/hashicorp/memberlist v0.1.0 // indirect
	github.com/hashicorp/serf v0.8.1 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/influxdata/influxdb v1.7.0 // indirect
	github.com/influxdata/platform v0.0.0-20181112180758-f643405ee645 // indirect
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515 // indirect
	github.com/kr/pty v1.1.3 // indirect
	github.com/miekg/dns v1.0.15 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/gomega v1.4.2 // indirect
	github.com/openzipkin/zipkin-go v0.1.3 // indirect
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/common v0.0.0-20181109100915-0b1957f9d949 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/smartystreets/goconvey v0.0.0-20181108003508-044398e4856c // indirect
	github.com/tv42/httpunix v0.0.0-20150427012821-b75d8614f926 // indirect
	go.opencensus.io v0.18.0 // indirect
	golang.org/x/crypto v0.0.0-20181106171534-e4dc69e5b2fd // indirect
	golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3 // indirect
	golang.org/x/net v0.0.0-20181108082009-03003ca0c849
	golang.org/x/oauth2 v0.0.0-20181106182150-f42d05182288
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f // indirect
	golang.org/x/sys v0.0.0-20181107165924-66b7b1311ac8 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	golang.org/x/tools v0.0.0-20181112162442-680468b7556f // indirect
	google.golang.org/api v0.0.0-20181108001712-cfbc873f6b93
	google.golang.org/appengine v1.3.0
	google.golang.org/genproto v0.0.0-20181109154231-b5d43981345b
	google.golang.org/grpc v1.16.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	honnef.co/go/tools v0.0.0-20180920025451-e3ad64cb4ed3 // indirect
)
