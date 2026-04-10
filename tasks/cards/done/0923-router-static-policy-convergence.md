# Card 0923: Router Static Policy Convergence

Priority: P1
State: done
Primary Module: router

## Goal

Keep stable `router` static support as a small mount primitive. Move or remove richer static-serving policy from `router` when it belongs to `x/frontend`, reference apps, or direct `http.FileServer` wiring.

## Problem

`router/static.go` currently mixes a simple static mount with policy-heavy behavior:

- local directory and `http.FileSystem` dispatch through a single `Root any` field
- cache-control policy
- ETag generation
- index-file policy
- SPA fallback behavior
- allowed-extension filtering
- file hash and file metadata handling

The stable router manifest allows "Static mount primitives", but not frontend asset policy. The current shape creates overlapping ownership:

- `router`: route structure and static mount primitive
- `x/frontend`: frontend asset serving and SPA/static behavior
- application/reference wiring: project-specific cache headers and fallback decisions

`Root any` is also less explicit than the rest of the stable API, which otherwise favors concrete stdlib types and grep-friendly control flow.

## Scope

- Decide the stable primitive shape for router static serving: local dir and/or `http.FileSystem`.
- Remove policy-heavy static fields from stable router if they are not required for the primitive.
- Move frontend/SPA-style behavior to `x/frontend` if a generic owner is needed.
- Replace `Root any` with explicit fields or constructors if static config remains.
- Update tests to assert path traversal protection and primitive static behavior only.
- Update docs and manifests to document the static boundary.

## Non-Goals

- Do not remove the ability to mount static files through router.
- Do not add external dependencies.
- Do not move route matching or params into `x/frontend`.
- Do not add application-specific cache policy to stable router.

## Expected Files

- `router/static.go`
- `router/static_test.go`
- `x/frontend/*` if policy is moved
- `docs/modules/router/README.md`
- `docs/modules/x-frontend/README.md`
- `router/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./router ./x/frontend
go test -race -timeout 60s ./router
go vet ./router ./x/frontend
```

Then run the required repo-wide gates before committing.

## Done Definition

- Stable router static API is explicit and primitive-level.
- SPA fallback, cache header policy, ETag policy, and allowed-extension policy no longer live in stable router unless explicitly justified in the final diff.
- Static path traversal and null-byte protections remain covered.
- Docs/manifests explain the boundary between `router` and `x/frontend`.
- Focused gates and repo-wide gates pass.

## Outcome

- Removed `router.StaticConfig` and `Router.StaticWithConfig`; no in-repo callers remained.
- Kept stable `router.Static` and `router.StaticFS` as explicit static mount primitives.
- Removed router-owned cache header policy, ETag generation, SPA fallback, index-file policy, and allowed-extension filtering.
- Documented that frontend asset policy belongs to `x/frontend`.
