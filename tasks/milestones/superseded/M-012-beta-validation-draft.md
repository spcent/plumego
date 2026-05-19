# M-012: Input Validation Bridge

- **Branch:** `milestone/M-012-input-validation`
- **Depends on:** M-008
- **Parallel OK:** yes

---

## Goal

Graduate `x/validate` from Experimental to Beta by implementing a `Validator` interface, a generic `Bind[T]` function, and structured validation error integration with `contract.WriteError`, so every plumego application has a first-class, explicit request-binding and validation path.

---

## Architecture Decisions

- `x/validate` must remain no-tag, no-reflection-based struct scanning. Callers call `Bind[T]` explicitly; no annotation processor.
- Validation errors map to `contract.APIError` with code `"validation_error"` and a `fields` detail map — no new error envelope type.
- `Bind[T]` reads JSON via `json.NewDecoder(r.Body).Decode` — the same pattern as CLAUDE.md.
- The `Validator` interface is a single-method interface: `Validate() error`. Types opt in by implementing it; no struct tags.
- `x/validate` is a separate Go module under `x/validate/go.mod`. It may import `contract` but not any other `x/*`.
- No external validation library dependency (e.g., no `go-playground/validator`). Stdlib only.
- After promotion, update `specs/extension-maturity.yaml` and `docs/EXTENSION_MATURITY.md`.

---

## Context — Read Before Touching Code

1. `AGENTS.md`
2. `specs/dependency-rules.yaml`
3. `specs/task-routing.yaml`
4. `x/validate/module.yaml`
5. `contract/errors.go`
6. `specs/change-recipes/new-extension-module.yaml`
7. `reference/standard-service/internal/app/handlers/`

## Affected Modules

- **Primary:** `x/validate`
- **Secondary:** `contract` (read-only reference), `specs/extension-maturity.yaml`, `docs/EXTENSION_MATURITY.md`

---

## Tasks

### Phase 1 — Orient (sequential)

1. Read every file in the **Context** section above.
2. Read all existing files in `x/validate/` to understand the current experimental surface.
3. Identify the exact change points; do not modify anything yet.

### Phase 2 — Implement (parallel)

- [ ] `x/validate/validate.go`: define `Validator` interface (`Validate() error`), `Bind[T]` generic function (decode then validate), `BindJSON[T]` alias for clarity.
- [ ] `x/validate/errors.go`: `ValidationError` type that wraps per-field errors and maps to a `contract.APIError` via `ToAPIError() *contract.APIError`.
- [ ] `x/validate/validate_test.go`: table-driven tests covering — valid payload binds without error, missing required field returns `ValidationError`, invalid JSON returns decode error, `Validator.Validate()` error surfaces in `ValidationError.Fields`.
- [ ] `x/validate/example_test.go`: runnable godoc example showing a handler using `Bind[T]` and `contract.WriteError`.
- [ ] `x/validate/module.yaml`: update `status` field to `beta`.
- [ ] `specs/extension-maturity.yaml`: promote `x/validate` status to `beta`.
- [ ] `docs/EXTENSION_MATURITY.md`: update maturity table row for `x/validate`.

### Phase 3 — Test (sequential)

8. Run `go test -race -timeout 60s ./x/validate/...`.
9. Confirm negative path: a struct that fails `Validate()` returns a non-nil `ValidationError` with non-empty `Fields`.
10. Confirm `gofmt -l ./x/validate/` is empty.

### Phase 4 — Validate and Ship (sequential)

11. Run the **Acceptance Criteria** commands; fix any failures.
12. Commit: `feat(x/validate): promote to beta with Bind[T] and Validator interface [M-012]`
13. Final commit: `milestone(M-012): Input Validation Bridge`
14. Push to `milestone/M-012-input-validation`.

---

## Acceptance Criteria

```bash
go test -race -timeout 60s ./x/validate/...
go vet ./x/validate/...
gofmt -l ./x/validate/
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/public-entrypoints-sync
```

Expected: all exit 0; `gofmt -l` outputs nothing.

---

## Out of Scope

- Do not add struct tag scanning or reflection-based validators.
- Do not introduce external validation library dependencies.
- Do not add `x/validate` to the root `go.mod`.
- Do not change `contract` package internals; read only.
- Do not add validation to any reference app in this milestone (that is a follow-up card).

---

## Open Questions

(none at spec time)

---

## Done Definition

- [ ] All Acceptance Criteria commands exit 0.
- [ ] `gofmt -l ./x/validate/` produces no output.
- [ ] `x/validate/module.yaml` status is `beta`.
- [ ] `specs/extension-maturity.yaml` reflects beta status.
- [ ] `docs/EXTENSION_MATURITY.md` maturity table updated.
- [ ] Godoc example in `example_test.go` is runnable.
- [ ] Branch `milestone/M-012-input-validation` pushed.
- [ ] PR open, title `milestone(M-012): Input Validation Bridge`.
