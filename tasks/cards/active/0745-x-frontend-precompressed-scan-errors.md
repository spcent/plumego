# Card 0745: x/frontend Precompressed Scan Errors

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/compression.go`
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0744

Goal:
Make directory-backed precompressed variant indexing fail visibly instead of
silently dropping entries on scan errors.

Scope:
- Return construction errors when directory variant scanning encounters
  filesystem errors.
- Preserve lazy best-effort probing for custom non-directory `http.FileSystem`
  implementations.
- Add focused coverage for scan error behavior where the platform supports it.
- Update docs with the fail-fast directory scan contract.

Non-goals:
- Do not add file watching or live reload.
- Do not scan arbitrary custom filesystems.
- Do not add logging.

Files:
- `x/frontend/compression.go`
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document that directory-backed precompressed metadata is fail-fast.

Done Definition:
- Directory-backed variant scan errors surface during mount construction.
- Custom filesystem lazy probing remains unchanged.
- The listed validation commands pass.
