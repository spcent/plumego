# 1234 - Core TLS Basic Surface Decision

State: done
Priority: P1
Primary module: core docs

## Goal

Freeze the stable decision that core owns only basic cert/key TLS loading, while advanced TLS policy remains caller-owned through the prepared `*http.Server`.

## Scope

- Clarify the TLS ownership decision in core docs and README guidance.
- Add public coverage proving callers can adjust the prepared server TLS config before serving.
- Avoid adding new `AppConfig` fields before v1.

## Non-goals

- Do not add a TLS callback or hook.
- Do not expose a custom `*tls.Config` field on `AppConfig`.
- Do not change certificate loading behavior.

## Files

- `core/core_public_test.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`
- `tasks/cards/done/1234-core-tls-basic-surface-decision.md`

## Tests

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Update core docs and bilingual README guidance.

## Done Definition

- Docs state that core's stable TLS API remains basic cert/key loading.
- Public tests show advanced TLS policy is caller-owned through `Server().TLSConfig`.

## Validation

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`
