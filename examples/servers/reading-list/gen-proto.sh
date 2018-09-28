#!/bin/sh

go get -u github.com/nytimes/openapi2proto/cmd/openapi2proto

openapi2proto -spec service.yaml -options > service.proto;

# for our code
protoc --go_out=plugins=grpc:. service.proto;

# for Cloud Endpoints
protoc --include_imports --include_source_info service.proto --descriptor_set_out service.pb
