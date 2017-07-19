all: test

deps:
	go get -d -v github.com/NYTimes/gizmo/...

updatedeps:
	go get -d -v -u -f github.com/NYTimes/gizmo/...

testdeps:
	go get -d -v -t github.com/NYTimes/gizmo/...

updatetestdeps:
	go get -d -v -t -u -f github.com/NYTimes/gizmo/...

build: deps
	go build github.com/NYTimes/gizmo/...

install: deps
	go install github.com/NYTimes/gizmo/...

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go\|\.pb\.gw\.go\|examples\|pubsub\/aws\/awssub_test\.go' | grep -v 'server\/kit\/kitserver_pb_test\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet github.com/NYTimes/gizmo/...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck -ignoretests github.com/NYTimes/gizmo/...

pretest: lint vet # errcheck

test: testdeps pretest
	go test github.com/NYTimes/gizmo/...

clean:
	go clean -i github.com/NYTimes/gizmo/...

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
	lint \
	vet \
	errcheck \
	pretest \
	test \
	clean \
	coverage
