# 0311 · x/mq — 20 sentinel errors buried in types.go

## Status
active

## Problem
`x/mq/types.go` contains the package's Go type definitions **and** 20 sentinel error
variables (lines 78–132). A separate `queue_errors.go` holds 4 more task-queue-specific
sentinels. Having error definitions split across two non-error files makes them hard to
discover, and mixing type declarations with error vars in types.go violates the single
responsibility of that file.

```
types.go:78   ErrRecoveredPanic
types.go:81   ErrNotInitialized
...           (18 more)
queue_errors.go:6  ErrDuplicateTask  (only 4 here)
queue_errors.go:7  ErrTaskNotFound
queue_errors.go:8  ErrLeaseLost
queue_errors.go:9  ErrTaskExpired
```

## Fix
1. Create (or repurpose) `x/mq/errors.go` as the **single** home for all sentinel errors.
2. Move all 20 vars from `types.go` into `errors.go`.
3. Move all 4 vars from `queue_errors.go` into `errors.go`; delete `queue_errors.go`.
4. Keep the same var names and message strings — no semantic change.
5. `types.go` becomes pure type/interface declarations.

## Verification
```bash
go build ./x/mq/...
go test ./x/mq/...
grep -n "errors\.New" x/mq/types.go   # should be empty
ls x/mq/queue_errors.go               # should not exist
```

## Affected files
- `x/mq/errors.go` (new)
- `x/mq/types.go` (remove error var block)
- `x/mq/queue_errors.go` (delete)

## Severity
Medium — discoverability and single-responsibility; no functional change
