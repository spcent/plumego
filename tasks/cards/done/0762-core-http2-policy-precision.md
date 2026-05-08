# 0762 - Core HTTP2 Policy Precision

State: done
Priority: P2
Primary module: core docs

## Goal

Make `HTTP2Enabled` precise as a prepared `http.Server` TLS HTTP/2 policy, not a universal HTTP/2 switch.

## Scope

- Clarify the config field comment.
- Clarify core docs and README guidance.
- Add or adjust tests that assert disabled policy installs the expected `TLSNextProto` override.

## Non-goals

- Do not rename `HTTP2Enabled`.
- Do not change default behavior.
- Do not add h2c support or protocol negotiation features.

## Files

- `core/config.go`
- `core/core_public_test.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`
- `tasks/cards/active/0762-core-http2-policy-precision.md`

## Tests

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Update core docs and bilingual README guidance.

## Done Definition

- `HTTP2Enabled` wording matches implementation.
- Tests cover the prepared server TLS HTTP/2 override.

## Validation

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`
