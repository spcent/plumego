#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d "${TMPDIR:-/tmp}/plumego-stable-api.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT

cd "$ROOT"

go run ./internal/checks/extension-api-snapshot \
	-module ./core \
	-out "$TMP/core-head.snapshot"

go run ./internal/checks/extension-api-snapshot \
	-compare \
	docs/stable-api/snapshots/core-head.snapshot \
	"$TMP/core-head.snapshot"
