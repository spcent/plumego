# Card 0984: Add Missing ErrorType Constants — MethodNotAllowed, NotImplemented, BadGateway

Priority: P3
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: contract

## Goal

The `contract` error type lookup table (`errors.go`) is missing three
`ErrorType` constants that are commonly needed by handlers.  Six call sites
currently spell out Status + Code + Category manually for 405 responses because
no `TypeMethodNotAllowed` type exists.

## New Types to Add in `contract/errors.go`

```go
// HTTP protocol errors
TypeMethodNotAllowed ErrorType = "method_not_allowed"
TypeNotImplemented   ErrorType = "not_implemented"
TypeBadGateway       ErrorType = "bad_gateway"
```

Metadata entries (`errorTypeLookup`):

```go
TypeMethodNotAllowed: {CategoryClient, CodeMethodNotAllowed, http.StatusMethodNotAllowed},
TypeNotImplemented:   {CategoryServer, CodeNotImplemented,   http.StatusNotImplemented},
TypeBadGateway:       {CategoryServer, CodeBadGateway,       http.StatusBadGateway},
```

## Call Sites to Simplify After Adding

### TypeMethodNotAllowed (×6 sites → 1 builder call each)

| File | Current pattern |
|---|---|
| `x/ai/streaming/handler.go:92` | `.Status(http.StatusMethodNotAllowed).Code(contract.CodeMethodNotAllowed).Category(contract.CategoryClient)` |
| `x/frontend/frontend.go:348` | same |
| `x/pubsub/distributed.go:749` | same |
| `x/pubsub/prometheus.go:510` | same |
| `x/scheduler/admin_http.go:351` | same |
| `x/websocket/server.go:183` | same |

Each becomes: `.Type(contract.TypeMethodNotAllowed)`

### TypeNotImplemented (×1 site)

| File | Current pattern |
|---|---|
| `x/rest/resource.go:106` | `.Status(http.StatusNotImplemented).Code(contract.CodeNotImplemented)` |

Becomes: `.Type(contract.TypeNotImplemented)`

### TypeBadGateway (×1 potential site)

| File | Current pattern |
|---|---|
| `x/messaging/api.go:153` | `.Status(http.StatusBadGateway).Category(contract.CategoryServer).Code("PROVIDER_ERROR")` |

Not a direct replacement (Code differs), but TypeBadGateway is still useful
for other callers that want the canonical 502 response.

## Tests

```bash
go test -timeout 20s ./contract/...
go test -timeout 20s ./x/rest/... ./x/websocket/... ./x/scheduler/... ./x/frontend/... ./x/ai/... ./x/pubsub/...
go vet ./contract/... ./x/...
```

## Done Definition

- `TypeMethodNotAllowed`, `TypeNotImplemented`, `TypeBadGateway` are defined
  in `contract/errors.go` with correct metadata entries.
- All 6 MethodNotAllowed call sites simplified to `.Type(contract.TypeMethodNotAllowed)`.
- `x/rest/resource.go` simplified to `.Type(contract.TypeNotImplemented)`.
- All tests pass; `go vet` clean.

## Outcome

Completed. Added `TypeMethodNotAllowed`, `TypeNotImplemented`, and `TypeBadGateway` to `contract/errors.go` with correct metadata entries. All 6 MethodNotAllowed call sites simplified to `.Type(contract.TypeMethodNotAllowed)`. `x/rest/resource.go` simplified to `.Type(contract.TypeNotImplemented)`. All tests pass; `go vet` clean.
