#!/bin/bash

mkdir -p bin
go build -o bin/main main.go api.go converstaion.go

if [[ ! -z "$1" ]]; then
    ./bin/main
fi
