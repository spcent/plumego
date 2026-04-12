# 0946 · x/ipc — 6 sentinel errors defined in ipc.go

## Status
active

## Problem
`x/ipc/ipc.go` defines the package's interfaces, types, and constants — **and** 6 sentinel
error variables in the middle of the file (lines 131–141):

```go
var (
    ErrServerClosed        = errors.New("ipc: server closed")
    ErrClientClosed        = errors.New("ipc: client closed")
    ErrInvalidConfig       = errors.New("ipc: invalid configuration")
    ErrConnectTimeout      = errors.New("ipc: connection timeout")
    ErrPlatformNotSupported = errors.New("ipc: platform not supported")
    ErrInvalidAddress      = errors.New("ipc: invalid address")
)
```

There is no `errors.go` file in the package. Mixing type declarations and error vars in
the same file makes both harder to scan.

## Fix
1. Create `x/ipc/errors.go` with the standard header and move all 6 vars there.
2. Remove the `var` block and `"errors"` import (if no longer needed) from `ipc.go`.

## Verification
```bash
go build ./x/ipc/...
go test ./x/ipc/...
grep -n "errors\.New" x/ipc/ipc.go  # should be empty
```

## Affected files
- `x/ipc/errors.go` (new)
- `x/ipc/ipc.go` (remove error var block)

## Severity
Low — file organisation only; no semantic change
