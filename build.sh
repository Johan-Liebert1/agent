#!/bin/bash

mkdir -p bin
go build -o bin/main src/*.go

if [[ ! -z "$1" ]]; then
    ./bin/main
fi
