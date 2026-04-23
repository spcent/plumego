# Card 2125: Gateway Proxy Discovery Lifecycle
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: high
State: active
Primary Module: x/gateway
Owned Files:
- x/gateway/proxy.go
- x/gateway/proxy_test.go
- docs/modules/x-gateway/README.md
Depends On: none

Goal:
Make the gateway proxy service-discovery watcher follow the proxy lifecycle. `NewE`
currently starts `watchServiceDiscovery` with `context.Background()`, while
`Close` only stops health checks and closes the transport pool. A proxy that is
constructed with discovery can leave a watcher goroutine alive after `Close`.

Scope:
- Give the proxy an explicit cancellation path for discovery watch work.
- Make `Close` cancel discovery watching before closing the transport pool.
- Add a focused test with a blocking discovery implementation that proves `Close`
  cancels the watch context and does not leak the watcher path.
- Document the lifecycle behavior in the x/gateway module docs if the current
  docs describe proxy cleanup or discovery watches.

Non-goals:
- Do not change the public `ServiceDiscovery` interface unless the existing
  signature cannot express cancellation.
- Do not add a global watcher registry or background manager.
- Do not change health-check semantics beyond ordering needed for clean shutdown.

Tests:
- go test -race -timeout 60s ./x/gateway/...
- go test -timeout 20s ./x/gateway/...
- go vet ./x/gateway/...

Docs Sync:
Update `docs/modules/x-gateway/README.md` only if lifecycle or close semantics
are user-visible in the docs.

Done Definition:
- `Proxy.Close` stops discovery watch work started by `NewE`.
- Discovery watch tests cover context cancellation and shutdown ordering.
- The x/gateway validation commands pass.
- No stable root imports any `x/*` package.

Outcome:
Implemented proxy-owned discovery cancellation and wait semantics. `Proxy.Close`
now cancels the watcher context, waits for the watcher to exit, then stops health
checks and closes transports. Added a blocking discovery test that verifies close
cancels the watch context and remains idempotent.

Validation:
- `go test -race -timeout 60s ./x/gateway/...`
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`
