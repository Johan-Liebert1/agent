#!/bin/bash

set -ex

cd "${0%/*}"

mkdir -p bin
go build -o bin/main -p 10 src/*.go

if [[ ! -z "$1" ]]; then
    ./bin/main "${@:1}"
fi
