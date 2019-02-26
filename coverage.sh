#!/bin/bash

set -e

workdir=.cover
profile="$workdir/cover.out"
mode=set

generate_cover_data() {
    rm -rf "$workdir"
    mkdir -p "$workdir"
    go test -vet all -covermode="$mode" -coverprofile="$profile" "$@"
}

show_cover_report() {
    go tool cover -${1}="$profile"
}

push_to_coveralls() {
    goveralls -coverprofile="$profile"
}

generate_cover_data ./...
show_cover_report func
case "$1" in
"")
    ;;
--html)
    show_cover_report html ;;
--coveralls)
    push_to_coveralls ;;
*)
    echo >&2 "error: invalid option: $1" ;;
esac
rm -rf "$workdir"
