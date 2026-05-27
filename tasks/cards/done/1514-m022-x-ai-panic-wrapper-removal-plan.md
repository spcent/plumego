# Card 1514

Milestone: M-022
Recipe: specs/change-recipes/docs-and-config.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/provider/manager.go`
- `x/ai/streaming/streaming.go`
- `x/ai/resilience/provider.go`
- `x/ai/semanticcache/provider.go`
- `x/ai/metrics/metrics.go`
- `docs/modules/x/ai/README.md`
- `specs/deprecation-inventory.yaml`
Depends On: 1511

Goal:
- Turn the verified panic-wrapper set in `x/ai` from undocumented compatibility
  behavior into explicit deprecated surfaces with tracked replacements and
  removal conditions.

Scope:
- Add `Deprecated:` guidance to the verified compatibility wrappers.
- Record each wrapper in `specs/deprecation-inventory.yaml` with a replacement
  and keep/remove decision.
- Update `docs/modules/x/ai/README.md` so the error-returning alternatives are
  the documented preferred path.

Non-goals:
- Do not break stable-tier API by removing `provider.Manager.Register` or
  `streaming.StreamManager.Register` in this card.
- Do not refactor all existing call sites to E-variants here.
- Do not touch unrelated non-panic constructors.

Files:
- `x/ai/provider/manager.go`
- `x/ai/streaming/streaming.go`
- `x/ai/resilience/provider.go`
- `x/ai/semanticcache/provider.go`
- `x/ai/metrics/metrics.go`
- `docs/modules/x/ai/README.md`
- `specs/deprecation-inventory.yaml`

Acceptance Tests:
- `go test -timeout 20s ./x/ai/...`

Tests:
- `go test -timeout 20s ./x/ai/...`
- `go run ./internal/checks/deprecation-inventory -strict`

Docs Sync:
- `docs/modules/x/ai/README.md`

Validation:
- `go test -timeout 20s ./x/ai/...`
- `go run ./internal/checks/deprecation-inventory -strict`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added explicit `Deprecated:` guidance to the verified `x/ai` panic wrappers:
  `provider.Manager.Register`, `streaming.StreamManager.Register`,
  `resilience.NewResilientProvider`, `semanticcache.NewSemanticCachingProvider`,
  and `metrics.Tags`.
- Registered those wrappers in `specs/deprecation-inventory.yaml` with concrete
  replacements and explicit keep decisions so strict inventory checks now track
  the compatibility surface instead of leaving it undocumented.
- Updated `docs/modules/x/ai/README.md` to steer callers toward `RegisterE`,
  `New*E`, and `TagsE` as the preferred non-panicking APIs.
- Validation:
  - `go test -timeout 20s ./x/ai/...`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `gofmt -l .`
