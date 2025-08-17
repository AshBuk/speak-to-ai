#!/usr/bin/env bash
set -euo pipefail

# Determine repository root and local libraries directory
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LIB_DIR="$REPO_ROOT/lib"

# Hint if libraries are not built yet
if [[ ! -f "$LIB_DIR/whisper.h" ]]; then
  echo "[dev-env] Whisper headers/libs not found in $LIB_DIR"
  echo "[dev-env] Please run: make whisper-libs"
fi

export CGO_ENABLED=1
export C_INCLUDE_PATH="$LIB_DIR${C_INCLUDE_PATH:+:$C_INCLUDE_PATH}"
export LIBRARY_PATH="$LIB_DIR${LIBRARY_PATH:+:$LIBRARY_PATH}"
export CGO_CFLAGS="-I$LIB_DIR ${CGO_CFLAGS:-}"
export CGO_LDFLAGS="-L$LIB_DIR -lwhisper -lggml-cpu -lggml ${CGO_LDFLAGS:-}"
export LD_LIBRARY_PATH="$LIB_DIR${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
export PKG_CONFIG_PATH="$LIB_DIR${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"

echo "[dev-env] CGO environment configured. lib: $LIB_DIR"
echo "[dev-env] You can now run: go build ./... | go test ./..."


