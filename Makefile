all: test

deps:
	go get -d -v github.com/nytimes/gizmo/...

updatedeps:
	go get -d -v -u -f github.com/nytimes/gizmo/...

testdeps:
	go get -d -v -t github.com/nytimes/gizmo/...

updatetestdeps:
	go get -d -v -t -u -f github.com/nytimes/gizmo/...

build: deps
	go build github.com/nytimes/gizmo/...

install: deps
	go install github.com/nytimes/gizmo/...

# to compile without installing protoc:
#   docker pull quay.io/pedge/protoeasy
#   docker run -d -p 6789:6789 quay.io/pedge/protoeasy
#   export PROTOEASY_ADDRESS=0.0.0.0:6789 # or whatever your docker host address is
proto:
	go get -v go.pedge.io/protoeasy/cmd/protoeasy
	protoeasy --grpc --go --go-import-path github.com/NYTimes/gizmo .

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go\|\.pb\.gw\.go\|examples\|pubsub\/awssub_test\.go\|pubsub\/pubsubtest'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet github.com/nytimes/gizmo/...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck -ignoretests github.com/nytimes/gizmo/...

pretest: lint vet errcheck

test: testdeps pretest
	go test github.com/nytimes/gizmo/...

clean:
	go clean -i github.com/nytimes/gizmo/...

coverage: testdeps
	./coverage.sh --coveralls

.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	install \
	proto \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	clean \
	coverage
