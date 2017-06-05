#!/bin/sh

openapi2proto -spec service.yaml -options > service.proto;

# for our code
protoc --go_out=plugins=grpc:. service.proto;

# for CE setup
protoc --include_imports --include_source_info service.proto --descriptor_set_out service.pb
