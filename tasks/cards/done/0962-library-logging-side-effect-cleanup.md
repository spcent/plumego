# Card 0962: Library Logging Side-Effect Cleanup

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: multi-module library ergonomics
Depends On: —

## Goal

Remove the remaining library-level stdout/stderr and package-global logging side
effects from reusable helper packages, so runtime behavior is caller-controlled
instead of being emitted implicitly by lower-level libraries.

## Problem

- `internal/nethttp/client.go` exports a `Logging` middleware that writes
  directly via package-global `log.Printf`, which hard-codes output policy into
  a reusable client helper.
- `x/ai/semanticcache/vectorstore/tiered.go` prints L2/L3 write failures with
  `fmt.Printf`, which writes directly to stdout from an experimental cache
  backend instead of returning structured information to the caller.
- `x/rest` and `x/websocket` have the same class of issue, but they are already
  covered by cards `0956` and `0960`; these two files are the remaining runtime
  code sites outside those cards.

## Scope

- Replace `internal/nethttp.Logging` with an explicit logging policy: injected
  logger, formatter callback, or removal in favor of caller-defined middleware.
- Replace `x/ai/semanticcache/vectorstore`'s `fmt.Printf` best-effort tier
  logging with explicit caller-visible behavior that does not write directly to
  stdout.
- Add tests for the chosen behavior where practical, especially for preserving
  best-effort semantics without hidden output side effects.
- Keep the change narrowly focused on runtime code, not README examples.

## Non-Goals

- Do not redesign retry policy, SSRF policy, or semantic-cache search logic.
- Do not reopen the `x/rest` or `x/websocket` logging cleanups already covered
  by other active cards.
- Do not add new external logging dependencies.

## Expected Files

- `internal/nethttp/client.go`
- `internal/nethttp/*_test.go`
- `x/ai/semanticcache/vectorstore/tiered.go`
- `x/ai/semanticcache/vectorstore/*_test.go`

## Validation

```bash
rg -n 'log\\.Printf\\(|fmt\\.Printf\\(' internal/nethttp x/ai/semanticcache/vectorstore -g '!**/*_test.go'
go test -timeout 20s ./internal/nethttp/... ./x/ai/semanticcache/vectorstore/...
go vet ./internal/nethttp/... ./x/ai/semanticcache/vectorstore/...
```

## Docs Sync

- only if package comments or examples must change because the logging helper
  API changes

## Done Definition

- `internal/nethttp` no longer hard-codes package-global log output in a public
  middleware helper.
- `x/ai/semanticcache/vectorstore` no longer writes tier failures directly to
  stdout.
- Best-effort semantics remain explicit and test-backed without hidden logging
  side effects.

## Outcome

- Replaced `internal/nethttp.Logging` with a caller-controlled callback middleware based on `RequestLogEntry`, removing package-global output policy from the helper.
- Replaced tiered vector-store `fmt.Printf` best-effort logging with an explicit `OnError` callback and `TieredBackgroundError` contract.
- Added focused tests for both callback surfaces so best-effort behavior remains explicit and covered without hidden stdout/stderr side effects.

## Validation Run

```bash
rg -n 'log\.Printf\(|fmt\.Printf\(' internal/nethttp x/ai/semanticcache/vectorstore -g '!**/*_test.go'
go test -timeout 20s ./internal/nethttp/... ./x/ai/semanticcache/vectorstore/...
go vet ./internal/nethttp/... ./x/ai/semanticcache/vectorstore/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```
