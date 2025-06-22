#!/bin/bash

set -ex

cd "${0%/*}"

mkdir -p bin
go build \
    -o bin/main \
    -p 10 \
    -ldflags "-X=main.APIKey=${API_KEY}" \
    src/*.go

if [[ ! -z "$1" ]]; then
    ./bin/main "${@:1}"
fi
