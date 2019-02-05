all: test

build:
	go build ./server/...
	go build ./pubsub/...
	go build ./auth/...
	go build ./config/...

lint:
	@cd /tmp; \
	go get -v golang.org/x/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go\|\.pb\.gw\.go\|examples\|pubsub\/aws\/awssub_test\.go' | grep -v 'server\/kit\/kitserver_pb_test\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet:
	go vet ./...

test: lint vet
	go test github.com/NYTimes/gizmo/...

clean:
	go clean -i github.com/NYTimes/gizmo/...

coverage:
	./coverage.sh --coveralls

.PHONY: \
	all \
	build \
	lint \
	vet \
	errcheck \
	test \
	clean \
	coverage
