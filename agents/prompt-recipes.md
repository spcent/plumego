# Prompt Recipes

Copy-paste task prompts for common shapes. Each includes a pre-filled preflight checklist.
For machine-readable recipes, see `specs/change-recipes/`.

---

## Recipe 1: Add an HTTP endpoint

```
Task: Add HTTP endpoint [METHOD] [PATH] to [module/handler file]

Preflight:
  Owning module: [e.g. reference/standard-service or x/rest]
  In-scope paths: [handler file, routes file, test file]
  Out-of-scope paths: core/, router/, contract/
  Public API impact: none (internal handler, not a package export)
  Security impact: none / [describe auth requirements]
  Validation plan: go test -race ./[module]/... && go vet ./[module]/...

Steps:
1. Read docs/CANONICAL_STYLE_GUIDE.md §3 (Application Structure)
2. Read reference/standard-service/internal/app/routes.go
3. Add one route registration line: app.[METHOD]("[PATH]", http.HandlerFunc([handler]))
4. Add handler with signature: func(w http.ResponseWriter, r *http.Request)
5. Use contract.WriteResponse for success, contract.WriteError for errors
6. Add test in [handler]_test.go using httptest.NewRecorder
7. Run: go test -race ./[module]/... && go vet ./[module]/...
```

---

## Recipe 2: Fix a middleware bug

```
Task: Fix [describe bug] in middleware/[sub-package]

Preflight:
  Owning module: middleware/[sub-package]
  In-scope paths: middleware/[sub-package]/*.go, middleware/[sub-package]/*_test.go
  Out-of-scope paths: core/, x/*, reference/
  Public API impact: none (fixing behavior, not signature)
  Security impact: [assess if auth/token middleware]
  Validation plan: go test -race ./middleware/... && go run ./internal/checks/dependency-rules

Steps:
1. Read middleware/module.yaml — check responsibilities and stop_conditions
2. Read middleware/[sub-package]/[file].go — understand current behavior
3. Fix the bug
4. Confirm: next is called exactly once; no business DTO assembly; no service injection
5. Add or update the failing test that reproduces the bug
6. Run: go test -race ./middleware/... && go vet ./middleware/...
7. Run: go run ./internal/checks/dependency-rules
```

---

## Recipe 3: Add an extension feature (x/*)

```
Task: Add [feature] to x/[family]

Preflight:
  Owning module: x/[family]
  In-scope paths: x/[family]/*.go, x/[family]/*_test.go
  Out-of-scope paths: core/, router/, contract/, middleware/ (stable roots)
  Public API impact: additive only (new exported symbol, no removal)
  Security impact: none / [describe]
  Validation plan: go test -race ./x/[family]/... && go run ./internal/checks/dependency-rules

Steps:
1. Read specs/task-routing.yaml — confirm x/[family] is the right destination
2. Read x/[family]/module.yaml — check allowed imports, stop_conditions
3. Read specs/dependency-rules.yaml entry for x/[family]
4. Implement in x/[family]/[feature].go
5. Add test in x/[family]/[feature]_test.go
6. If new exported symbol: add to x/[family]/module.yaml:public_entrypoints
7. Run: go test -race ./x/[family]/... && go vet ./x/[family]/...
8. Run: go run ./internal/checks/dependency-rules
```

---

## Recipe 4: Add a security primitive

```
Task: Add [security primitive] to security/[sub-package]

Preflight:
  Owning module: security/[sub-package]
  In-scope paths: security/[sub-package]/*.go, security/[sub-package]/*_test.go
  Out-of-scope paths: x/*, store/db, store/cache, core/
  Public API impact: additive only
  Security impact: yes — timing-safe comparison required, fail-closed required
  Validation plan: go test -race ./security/... && go run ./internal/checks/dependency-rules

Steps:
1. Read security/module.yaml — verify sub-package scope
2. Read specs/dependency-rules.yaml security entry
3. Implement — use crypto/subtle for secret comparison, return error on any failure
4. Add negative-path tests: invalid input, empty input, malformed input
5. Confirm: no sensitive values in log output, fail-closed on all error paths
6. Run: go test -race ./security/... && go vet ./security/...
```

---

## Recipe 5: Update a module manifest

```
Task: Update module.yaml for [module]

Preflight:
  Owning module: [module]
  In-scope paths: [module]/module.yaml
  Out-of-scope paths: everything else
  Public API impact: none (manifest is not Go code)
  Security impact: none
  Validation plan: go run ./internal/checks/module-manifests

Steps:
1. Read specs/module-manifest.schema.yaml — check required fields
2. Edit [module]/module.yaml — update only the fields that changed
3. Run: go run ./internal/checks/module-manifests
4. Run: go run ./internal/checks/agent-workflow
```

---

## Recipe 6: Write a contract/freeze test

```
Task: Add freeze test for [module] exported API

Preflight:
  Owning module: [module]
  In-scope paths: [module]/freeze_test.go (create or update)
  Out-of-scope paths: all other modules
  Public API impact: none (test only)
  Security impact: none
  Validation plan: go test -run TestFreeze ./[module]/...

Steps:
1. Read [module]/module.yaml:public_entrypoints — enumerate expected exported symbols
2. Create [module]/freeze_test.go in package [module]
3. Write TestFreeze[Module]PublicSurface:
   - Use go/parser or reflect to enumerate exported names
   - Assert the set matches the expected list
   - Fail with a clear message listing any unexpected additions or removals
4. Run: go test -run TestFreeze ./[module]/...
5. This test will fail when any exported symbol is added or removed — that is intentional
```

---

## Recipe 7: Add a stable root change (high risk)

```
Task: [Describe change] in [stable-root package]

Preflight:
  Owning module: [stable-root]
  In-scope paths: [specific files only]
  Out-of-scope paths: all x/*, reference/*, other stable roots
  Public API impact: [explicit description or "none"]
  Security impact: [explicit description or "none"]
  Validation plan: make check-boundaries && make test-stable

Steps:
1. Read AGENTS.md §2 (Non-Negotiables) — confirm this change is allowed
2. Read [module]/module.yaml:stop_conditions — confirm none apply
3. Read docs/architecture/core-boundary.md or extension-boundary.md if relevant
4. Make the smallest possible change
5. Update [module]/module.yaml:public_entrypoints if an exported symbol changed
6. Update freeze_test.go if the exported set changed
7. Run: make check-boundaries && make test-stable
8. Document any residual risk in the PR description
```
