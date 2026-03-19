#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

export CONFIG_FILE=example/config.yaml
export SECRET_FILE=example/secret.yaml

if ! command -v go &> /dev/null; then
    echo "[ERROR] go command not found" >&2
    exit 1
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "[ERROR] Config file not found: $CONFIG_FILE" >&2
    exit 1
fi

if [ ! -f "$SECRET_FILE" ]; then
    echo "[ERROR] Secret file not found: $SECRET_FILE" >&2
    exit 1
fi

echo "[INFO] Starting app..." >&2
go run ./cmd/app
