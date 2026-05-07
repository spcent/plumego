# 0724 - x/websocket doc manifest API inventory

Status: active
Priority: P1
Primary module: `x/websocket`

## Problem

`module.yaml` and the primer list only a small subset of exported symbols while
the package exposes many more types and helpers. Stable readiness requires a
truthful public API inventory and examples that match current constructors and
route wiring.

## Scope

- Generate or manually audit exported `x/websocket` symbols.
- Update `module.yaml` and `docs/modules/x-websocket/README.md` to match the
  intended public surface.
- Remove stale examples and maturity language drift.
- Keep `experimental` status unless governance evidence changes separately.

## Out of Scope

- API removal beyond documentation truth.
- Promotion to beta/stable.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/module-manifests`

