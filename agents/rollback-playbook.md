# Rollback Playbook

How to cleanly revert each change type. Use this when a change needs to be undone
to avoid partial-revert states that break tests or checks.

---

## Principle

Always revert in reverse dependency order: callers before definitions.

---

## New Exported Symbol

**Rollback steps:**
1. `rg -n 'SymbolName' . --include='*.go'` — list every site
2. Remove or replace callers in reverse dependency order (leaf packages first)
3. Remove the symbol definition
4. Remove from `module.yaml:public_entrypoints`
5. Remove from `freeze_test.go` expected set (if present)

**Verification:**
```bash
go build ./...
go run ./internal/checks/public-entrypoints-sync
```
`rg 'SymbolName' . --include='*.go'` must return zero results.

---

## New Middleware Sub-Package

**Rollback steps:**
1. Remove all registration of the middleware in reference apps or tests
2. Delete the sub-package directory
3. Remove from `middleware/module.yaml` if declared

**Verification:**
```bash
go test ./middleware/...
go run ./internal/checks/module-manifests
```

---

## New module.yaml Field

**Rollback steps:**
1. Remove the field from the module.yaml file
2. If the field was added to `specs/module-manifest.schema.yaml`, remove it there too

**Verification:**
```bash
go run ./internal/checks/module-manifests
go run ./internal/checks/agent-workflow
```

---

## New Dependency

**Rollback steps:**
1. Remove the import from all `.go` files
2. `go mod tidy`

**Verification:**
```bash
git diff go.mod go.sum
```
Both files must be unchanged from before the dependency was added.

---

## Route Registration Change

**Rollback steps:**
1. Revert the route line in `routes.go` (or equivalent)
2. Revert or delete the handler function
3. Revert or delete the handler test

**Verification:**
```bash
go test -race ./[app]/...
```
No 404 or routing regression.

---

## Security Primitive Change

**Rollback steps:**
1. Revert the security function/type definition
2. Revert any callers (middleware, handlers, tests) in reverse order
3. Check that no test relies on the old behavior

**Verification:**
```bash
go test -race ./security/...
go run ./internal/checks/dependency-rules
```
All negative-path tests must still pass.

---

## Stable Root Behavior Change

**Rollback steps:**
1. `git diff HEAD~1 -- [module]/` — list all changed files
2. Revert callers in x/* and reference/* first (if any)
3. Revert the stable root changes
4. If `freeze_test.go` was updated, revert it too

**Verification:**
```bash
make check-boundaries
make test-stable
```

---

## Module.yaml:public_entrypoints Change

**Rollback steps:**
1. Revert the module.yaml change
2. Run `go run ./internal/checks/public-entrypoints-sync` — it will report drift if the Go code still has the symbol

**Verification:**
```bash
go run ./internal/checks/public-entrypoints-sync
```

---

## Spec File Change (specs/*.yaml)

**Rollback steps:**
1. `git revert` the spec file commit
2. Run all boundary checks to confirm no residual state

**Verification:**
```bash
make check-boundaries
```

---

## Git Commands for Quick Revert

```bash
# Revert a single file to HEAD
git checkout HEAD -- path/to/file.go

# Revert all changes in a directory to HEAD
git checkout HEAD -- path/to/dir/

# Show what a revert would affect
git diff HEAD~1 -- path/to/file.go

# Interactive revert of last commit (do NOT use on shared branches)
git revert HEAD
```
