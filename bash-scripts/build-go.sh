#!/bin/bash

# Build Go application with CGO
set -e

export C_INCLUDE_PATH="$(pwd)/lib"
export LIBRARY_PATH="$(pwd)/lib"
export CGO_CFLAGS="-I$(pwd)/lib"
export CGO_LDFLAGS="-L$(pwd)/lib -lwhisper -lggml-cpu -lggml"
export LD_LIBRARY_PATH="$(pwd)/lib:$LD_LIBRARY_PATH"
export PKG_CONFIG_PATH="$(pwd)/lib:$PKG_CONFIG_PATH"

if [ "$1" = "test" ]; then
    go test "${@:2}"
else
    go build -v -o speak-to-ai cmd/daemon/main.go
fi 