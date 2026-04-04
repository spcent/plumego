# Card 0606

Priority: P1

Goal:
- Remove the deprecated package-level `contract.Param(r, key)` function by
  migrating its two remaining callers in the `router` package, then deleting
  the function.

Problem:
- `contract.Param` (context_core.go:206) is marked `// Deprecated` (added in
  card 0605) but still has two live callers:
  - `router/router.go:413` — `router.Param(r, name)` wraps `contract.Param`
  - `router/static.go:58` — `getFilePathFromRequest` calls `contract.Param`
- Style guide §16.5: deprecated symbols are removed in the same PR that
  replaces their last caller; dead wrappers must not remain.
- The function duplicates `Ctx.Param` and `contract.ParamsFromContext`, which
  are the canonical ways to access route parameters.

Migration:
- `router/router.go:413`:
  ```go
  // Before
  v, _ := contract.Param(r, name)
  // After
  v, _ := contract.ParamsFromContext(r.Context())[name]  // or via Ctx.Param
  ```
  Router's own `Param(r, name) string` can delegate to
  `contract.ParamsFromContext(r.Context())` directly.

- `router/static.go:58`:
  ```go
  // Before
  relPath, ok := contract.Param(req, "filepath")
  // After
  params := contract.ParamsFromContext(req.Context())
  relPath, ok := params["filepath"]
  ```

After both callers are migrated, delete `contract.Param` (context_core.go:204-217).

Non-goals:
- Do not change `ParamsFromContext` or any other accessor.
- Do not change `Ctx.Param` or `Ctx.MustParam`.

Files:
- `contract/context_core.go` (delete the function)
- `router/router.go` (migrate caller)
- `router/static.go` (migrate caller)

Tests:
- `go test ./contract/... ./router/...`
- `go build ./...`
- `grep -rn 'contract\.Param\b'` must return empty

Done Definition:
- `contract.Param` does not exist.
- `router.Param(r, name)` delegates directly to `ParamsFromContext` without
  going through `contract.Param`.
- All tests pass.
