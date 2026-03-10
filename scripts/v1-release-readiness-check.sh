#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

RUN_RACE="${RUN_RACE:-0}"

SUBMODULES=(
  "cmd/plumego"
  "examples/crud-demo"
  "examples/db-metrics"
  "examples/file-storage-tests"
  "examples/multi-tenant-saas"
)

echo "[v1-check] Root formatting check"
root_fmt="$(gofmt -l .)"
if [ -n "$root_fmt" ]; then
  echo "[v1-check] gofmt check failed at root:"
  echo "$root_fmt"
  exit 1
fi

echo "[v1-check] Root tests"
go test -timeout 20s ./...

if [ "$RUN_RACE" = "1" ]; then
  echo "[v1-check] Root race tests"
  go test -race -timeout 60s ./...
fi

echo "[v1-check] Root vet"
go vet ./...

echo "[v1-check] Canonical doc drift check"
bash scripts/check-doc-api-drift.sh

for module in "${SUBMODULES[@]}"; do
  echo "[v1-check] Submodule: $module"
  (
    cd "$module"
    module_fmt="$(gofmt -l .)"
    if [ -n "$module_fmt" ]; then
      echo "[v1-check] gofmt check failed in $module:"
      echo "$module_fmt"
      exit 1
    fi
    go test -timeout 20s ./...
    go vet ./...
  )
done

echo "[v1-check] PASS"
