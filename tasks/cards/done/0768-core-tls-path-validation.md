# 0768 - Core TLS Path Validation

State: done
Priority: P1
Primary module: core

## Goal

Treat whitespace-only TLS cert/key paths as missing configuration instead of leaking lower-level file errors.

## Scope

- Trim-space validate `TLSConfig.CertFile` and `TLSConfig.KeyFile` when TLS is enabled.
- Preserve caller-provided paths for actual certificate loading.
- Add focused tests for whitespace cert/key paths.

## Non-goals

- Do not normalize paths before loading certificates.
- Do not add advanced TLS policy fields.
- Do not change TLS ownership decisions.

## Files

- `core/http_handler.go`
- `core/lifecycle_test.go`
- `tasks/cards/active/0768-core-tls-path-validation.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

No docs change expected unless validation wording changes.

## Done Definition

- TLS enabled with blank or whitespace cert/key paths fails through the stable missing-cert/key validation path.
- Focused core tests pass.

## Validation

- `gofmt -w core/http_handler.go core/lifecycle_test.go`
- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
