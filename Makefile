all: test

deps:
	go mod download

build: deps
	go build ./...

install: deps
	go install ./...

lint: deps
	GO111MODULE=off go get golang.org/x/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go\|\.pb\.gw\.go\|examples\|pubsub\/aws\/awssub_test\.go' | grep -v 'server\/kit\/kitserver_pb_test\.go'); do \
		golint -set_exit_status $${file}; \
	done

pretest: lint

test: deps pretest
	go test -vet all ./...

coverage: deps
	./coverage.sh --coveralls

.PHONY: \
	all \
	deps \
	build \
	install \
	lint \
	pretest \
	test \
	coverage
