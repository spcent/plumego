# Verification Matrix

Exact gate commands per change type. Use the lightest sufficient profile.
Full profiles are defined in `specs/checks.yaml` and `specs/agent-quality-rules.yaml`.

---

## Gate Profiles

### Docs only
No Go gates required. Check that links resolve and terminology matches the code.

### Single module behavior change

```bash
go test -race -timeout 60s ./<module>/...
go test -timeout 20s ./<module>/...
go vet ./<module>/...
gofmt -w <changed-go-files>
```

### New middleware sub-package

```bash
go test -race -timeout 60s ./middleware/...
go vet ./middleware/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests
```

Also confirm: `next` called exactly once, conformance test passes.

### `x/*` behavior change

```bash
go test -race -timeout 60s ./x/<family>/...
go vet ./x/<family>/...
go run ./internal/checks/dependency-rules
```

### Stable root change (core / router / contract / middleware / security / store / health / log / metrics)

```bash
go test -race -timeout 60s ./<module>/...
go vet ./<module>/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
```

Or use `make check-boundaries && make test-stable`.

### Cross-module change (touches 2+ stable roots or stable + x/*)

```bash
make check-boundaries
make test-stable
go test -race -timeout 60s ./...
go vet ./...
gofmt -l .
```

### Public API change (new or removed exported symbol)

```bash
make gates-go
go run ./internal/checks/public-entrypoints-sync
go run ./internal/checks/deprecation-inventory -strict
```

Also: update `module.yaml:public_entrypoints`, update `freeze_test.go` if applicable.

### Security primitive change

```bash
go test -race -timeout 60s ./security/...
go vet ./security/...
go run ./internal/checks/dependency-rules
```

Confirm: timing-safe comparison used, fail-closed, negative-path tests present.

### New dependency — **STOP**

```
This requires human approval. Do not proceed.
Record the required package and reason in the task card's Open Questions section.
Create tasks/cards/blocked/<id>-dependency.md.
```

---

## Convenience Makefile Targets

| Target | What it runs |
|---|---|
| `make check-boundaries` | All 5 boundary checks, no tests |
| `make check-public-api` | Public-entrypoints-sync + API snapshot |
| `make test-stable` | Race tests for all stable roots |
| `make test-contract` | Only freeze/conformance/contract tests |
| `make bench` | All benchmarks, short mode |
| `make gates-go` | Full Go gates without pnpm website build |
| `make gates-website` | pnpm sync + check + build only |
| `make gates` | Full suite (gates-go + gates-website) |

---

## Module-Specific Focus

| Module | Extra tests to confirm |
|---|---|
| `router` | static, param, group, method, reverse-routing tests |
| `middleware` | ordering, error path, panic, timeout, cancel tests |
| `security` | invalid token, invalid signature, fail-closed, timing-safe tests |
| `x/tenant` | quota, policy, session isolation, negative-path tests |
| `contract` | JSON wire format, error code stability, freeze test |
| `core` | lifecycle order, startup, shutdown regression |
