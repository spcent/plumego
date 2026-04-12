# Card 0833

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/context_bind.go`
- `contract/context_extended_test.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `x/fileapi`

Goal:
- Shrink stable `contract.Ctx` back to transport binding and response responsibilities by removing multipart and disk-write upload helpers from the stable root.
- Converge file-upload convenience onto the owning extension instead of leaving filesystem-facing helpers inside transport contracts.

Problem:
- `contract/context_bind.go` currently exports both `Ctx.FormFile` and `Ctx.SaveUploadedFile`.
- `SaveUploadedFile` creates directories and writes uploaded content to disk, which is filesystem behavior rather than a transport contract primitive.
- `FormFile` and `SaveUploadedFile` currently have no production callers outside `contract` self-tests, so they enlarge the stable surface without carrying meaningful stable ownership.
- `contract/module.yaml` and `docs/modules/contract/README.md` describe request/response contracts, not multipart upload workflows or disk persistence helpers.

Scope:
- Remove multipart/file-upload convenience helpers from stable `contract.Ctx`.
- Keep `BindJSON`, `BindQuery`, validation, and response/error helpers as the canonical stable transport surface.
- Move any higher-level upload convenience that still needs a repo home to `x/fileapi`, where multipart HTTP ownership already lives.
- Update tests and any downstream callers in the same change.
- Sync the contract manifest and primer to the narrowed `Ctx` boundary.

Non-goals:
- Do not redesign `WriteResponse`, `WriteError`, bind errors, or validation helpers.
- Do not introduce a new stable file-upload abstraction in `contract`.
- Do not preserve deleted upload helpers via compatibility wrappers.
- Do not move application-specific storage policy into `x/fileapi`.

Files:
- `contract/context_bind.go`
- `contract/context_extended_test.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `x/fileapi`

Tests:
- `go test -timeout 20s ./contract/... ./x/fileapi/...`
- `go test -race -timeout 60s ./contract/... ./x/fileapi/...`
- `go vet ./contract/... ./x/fileapi/...`

Docs Sync:
- Keep the contract manifest and primer aligned on the rule that stable `contract` owns request/response contracts and bind helpers, while multipart upload handling belongs in `x/fileapi`.

Done Definition:
- Stable `contract.Ctx` no longer exports multipart upload or disk-save helpers.
- Any remaining upload convenience lives under `x/fileapi`.
- Contract tests and downstream callers are updated with no residual references to deleted `Ctx.FormFile` or `Ctx.SaveUploadedFile` APIs.
- Contract docs and manifest describe the same narrowed stable boundary the code implements.

Outcome:
- Completed.
- Removed multipart and file-upload convenience helpers from stable `contract.Ctx`, including file-writing helpers that exceeded the transport contract boundary.
- Upload parsing and file handling now stay in the owning transport/extension layers.
