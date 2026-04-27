# Card 0211

Priority: P1
State: done
Primary Module: middleware
Owned Files:
- `middleware/versioning/version.go`
- `middleware/transform/transform.go`
- `middleware/error_registry.go`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
Depends On:

Goal:
- Remove protocol- and resource-adaptation ownership from stable `middleware` so the stable root keeps only transport primitives.
- Converge directly to owning extension packages for API version negotiation and protocol transformation instead of leaving them under stable middleware.

Problem:
- Stable `middleware` still owns `versioning` and `transform`, which are resource/API protocol adaptation concerns rather than transport primitives.
- Repo routing guidance sends resource API standardization to `x/rest` and gateway/edge adaptation to `x/gateway`, but stable middleware still exports these capabilities directly.
- `middleware/error_registry.go` also still owns extension-specific transport codes such as `unsupported_version`, `transform_failed`, `protocol_transform_failed`, and `protocol_execution_failed`, which extends the stable boundary to cover extension behavior.

Scope:
- Relocate API version negotiation middleware to its owning extension package.
- Relocate request/response transformation middleware to its owning extension package.
- Remove extension-specific transport error codes from stable middleware and keep only generic transport errors in the stable root.
- Update stable and extension callers to the relocated surfaces in the same change.
- Sync the middleware manifest and primer to the post-pruning transport-only boundary.

Non-goals:
- Do not redesign stable primitives such as recovery, timeout, request ID, tracing, access logging, CORS, compression, or basic auth adapters.
- Do not move stable transport observability primitives out of `middleware`.
- Do not preserve forwarding wrappers in stable middleware for relocated protocol adapters.
- Do not turn `middleware` into a generic compatibility bucket for extension-layer edge features.

Files:
- `middleware/versioning/version.go`
- `middleware/transform/transform.go`
- `middleware/error_registry.go`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/... ./x/gateway/... ./x/rest/... ./x/tenant/...`
- `go test -race -timeout 60s ./middleware/... ./x/gateway/... ./x/rest/... ./x/tenant/...`
- `go vet ./middleware/... ./x/gateway/... ./x/rest/... ./x/tenant/...`

Docs Sync:
- Keep the middleware primer and manifest aligned on the rule that stable `middleware` owns transport primitives only, while protocol/resource adaptation lives in owning extensions.

Done Definition:
- Stable `middleware` no longer owns API version negotiation or protocol transformation middleware.
- Extension-specific transport error codes are removed from the stable root.
- Relocated protocol adapter surfaces live in owning extensions with updated callers.
- Stable middleware docs and manifest describe the same reduced transport-only boundary the code implements.
- No deprecated forwarding surface remains in stable middleware for removed protocol adapters or codes.

Outcome:
- Completed.
- Moved protocol/resource adaptation helpers out of stable `middleware`, including versioning and transform ownership, into their extension owners.
- Stable middleware now stays transport-only and no longer carries extension-specific error or adaptation surfaces.
