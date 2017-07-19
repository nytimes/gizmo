#!/bin/sh

export GCP_PROJECT_ID=local
export DATASTORE_EMULATOR_HOST=localhost:8082

gcloud beta emulators datastore start --host-port=localhost:8082 &

go run main.go
