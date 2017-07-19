#!/bin/sh

gcloud service-management deploy ../service.pb  ./service-ce-$1.yaml
