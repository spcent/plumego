# Card 0762

Milestone: M-003
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: todo
Primary Module: x/websocket
Owned Files:
- `x/websocket/module.yaml`
- `x/websocket/*.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
Depends On: 0761

Goal:
- Make the `x/websocket` stable-candidate API surface explicit and consistent with the module manifest.

Problem:
`module.yaml` lists only a small set of public entrypoints, but the package exports additional auth, validation, security, event, hub, and helper symbols. Stable cannot proceed while the intended compatibility surface is ambiguous.

Scope:
- Enumerate every exported symbol with `rg -n --glob '*.go'`.
- Classify each export as stable public, internal candidate, or remove-before-stable.
- Update `module.yaml public_entrypoints` to match the intended stable surface.
- Internalize or delete exports that are not part of the stable transport contract.
- Remove unused API noise such as saved-but-unused server/logger/debug state and dead hub metrics fields.

Non-goals:
- Do not change protocol behavior except where required by symbol removal.
- Do not introduce compatibility wrappers.
- Do not promote module maturity.

Files:
- `x/websocket/module.yaml`
- `x/websocket/*.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for every public symbol kept, renamed, or removed.

Done Definition:
- `module.yaml public_entrypoints` matches the real stable-candidate export list.
- Every remaining exported symbol has a clear stable purpose and Go doc.
- Removed or renamed symbols have no remaining Go call sites.

Outcome:
-
