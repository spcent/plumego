# 0944 · x/pubsub — extension sentinel errors bypass the package's structured *Error type

## Status
active

## Problem
`x/pubsub/errors.go` defines a well-designed structured error type:

```go
type Error struct { Code ErrorCode; Op string; Topic string; Message string; Cause error }
```

However, ten extension files each define their **own plain `errors.New()` sentinel
variables** that have nothing to do with this system:

| file | example sentinels |
|------|-------------------|
| `audit.go` | `ErrAuditClosed`, `ErrAuditCorrupted`, `ErrInvalidAuditQuery` |
| `backpressure.go` | `ErrBackpressureActive`, `ErrBackpressureClosed` |
| `consumergroup.go` | `ErrGroupClosed`, `ErrGroupNotFound`, `ErrConsumerNotFound`, … |
| `distributed.go` | `ErrClusterNotJoined`, `ErrNodeNotFound`, `ErrBroadcastFailed`, … |
| `dlq.go` | `ErrDLQClosed`, `ErrDLQNotFound`, `ErrInvalidQuery` |
| `multitenant.go` | `ErrQuotaExceeded`, `ErrTenantNotFound`, … |
| `ordering.go` | `ErrOrderingClosed`, `ErrInvalidOrderLevel`, … |
| `persistence.go` | `ErrPersistenceClosed`, `ErrInvalidWALEntry`, … (8 vars) |
| `ratelimit.go` | `ErrRateLimitExceeded`, `ErrInvalidRateLimit` |
| `replay.go` | `ErrReplayClosed`, `ErrMessageNotFound`, … |

Callers cannot use the structured `*Error` fields (Op, Topic, Cause) when catching these
errors, and the extension sentinels are scattered across implementation files rather than
being discoverable in one place.

Additionally, **generic names** like `ErrNodeNotFound` (in distributed.go) shadow the same
name in other packages — e.g. `x/mq` defines `ErrNodeNotFound` for a completely different
concept.

## Fix
Centralise: move all extension sentinel `var Err…` declarations into `errors.go`.
Keep the existing `*Error` structured type and `ErrCode` constants untouched.
The sentinel vars remain plain `errors.New()` (consistent with the rest of the codebase);
this is a file-organisation fix, not a type-system refactor.

After moving, the scattered files contain only type and method definitions — no `var` blocks.

1. Append all extension sentinels to `errors.go` (grouped by extension, with comments).
2. Remove the `var` blocks from the ten extension files.
3. Rename any generic names that clash across packages:
   - `ErrNodeNotFound` (distributed.go) → keep as-is (package-qualified on import anyway,
     but worth noting in comments)
   - `ErrInvalidQuery` (dlq.go) clashes with `ErrInvalidQuery` (replay.go: same name,
     different file, same package) — one of these is a duplicate; merge into one sentinel.

## Verification
```bash
go build ./x/pubsub/...
go test ./x/pubsub/...
grep -rn "^var Err\|^\tErr" x/pubsub/ --include="*.go" | grep -v "errors.go"  # should be empty
```

## Affected files
- `x/pubsub/errors.go` (append ~40 sentinel vars)
- `x/pubsub/audit.go`, `backpressure.go`, `consumergroup.go`, `distributed.go`,
  `dlq.go`, `multitenant.go`, `ordering.go`, `persistence.go`, `ratelimit.go`, `replay.go`
  (remove `var Err…` blocks)

## Severity
Medium — discoverability; ErrInvalidQuery name collision in same package is a latent bug
