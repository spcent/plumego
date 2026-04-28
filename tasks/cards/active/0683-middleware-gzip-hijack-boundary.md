# Card 0683

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
Depends On: 0682

Goal:
Preserve `http.Hijacker` support through gzip middleware before response compression starts, so transport-level connection upgrades and custom protocols do not break merely because the client sent `Accept-Encoding: gzip`.

Scope:
- Implement `Hijack` on the gzip response writer by forwarding to the underlying writer when supported.
- Mark headers as flushed/response started once the connection is hijacked so gzip finalization cannot write to the hijacked connection.
- Add tests for supported and unsupported hijack paths.
- Preserve existing gzip skip behavior for websocket upgrade requests.

Non-goals:
- Do not support hijacking after gzip compression has started.
- Do not add protocol transformation or gateway behavior.
- Do not redesign gzip buffering.

Files:
- `middleware/compression/gzip.go`
- `middleware/compression/gzip_test.go`

Tests:
- `go test -timeout 20s ./middleware/compression`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this preserves an expected `net/http` optional interface.

Done Definition:
- Handlers can observe and use `http.Hijacker` through gzip when the underlying writer supports it.
- Unsupported underlying writers return `http.ErrNotSupported`.
- Targeted middleware tests and vet pass.

Outcome:
