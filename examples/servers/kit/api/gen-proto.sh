#!/bin/sh

openapi2proto -spec service.yaml > service.proto;

protoc --go_out=plugins=grpc:. service.proto;
